package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/jobs"
	"github.com/yolo-hq/yolo/core/service"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/events"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/helpers"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

// CompleteRunAction records run completion and drives the task/PRD state machine.
type CompleteRunAction struct {
	action.TypedInput[inputs.CompleteRunInput]
	action.PublicAccess
	JobClient   *jobs.Client
	WorkflowJob jobs.Handler
}

func (a *CompleteRunAction) Description() string { return "Record run completion and advance state machine" }

func (a *CompleteRunAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	run, r := action.FindOrFail[entities.Run](ctx, action.ReadRepo[entities.Run](actx), actx.EntityID)
	if r != nil {
		return *r
	}

	input := a.Input(actx)
	now := time.Now()

	// 1. Update the run with completion data.
	_, err := action.Write[entities.Run](actx).Exec(ctx, write.Update{
		ID: actx.EntityID,
		Set: write.Set{
			fields.Run.Status.Value(input.Status),
			fields.Run.CostUSD.Value(input.CostUSD),
			fields.Run.TokensIn.Value(input.TokensIn),
			fields.Run.TokensOut.Value(input.TokensOut),
			fields.Run.DurationMs.Value(input.DurationMS),
			fields.Run.NumTurns.Value(input.NumTurns),
			fields.Run.Error.Value(input.Error),
			fields.Run.CommitHash.Value(input.CommitHash),
			fields.Run.FilesChanged.Value(input.FilesChanged),
			fields.Run.Result.Value(input.Result),
			fields.Run.SessionID.Value(input.SessionID),
			fields.Run.CompletedAt.Value(&now),
		},
	})
	if err != nil {
		return action.Failure(err.Error())
	}

	// 2. Load the parent task.
	task, err := action.ReadRepo[entities.Task](actx).FindOne(ctx, entity.FindOneOptions{ID: run.TaskID})
	if err != nil || task == nil {
		return action.Failure("failed to load task for run")
	}

	// 3. Dispatch based on run outcome.
	taskRead := action.ReadRepo[entities.Task](actx)
	taskWrite := action.WriteRepo[entities.Task](actx)
	prdWrite := action.WriteRepo[entities.PRD](actx)

	switch input.Status {
	case entities.RunCompleted:
		a.handleCompleted(ctx, actx, task, input, taskRead, taskWrite, prdWrite, now)
	case entities.RunFailed:
		a.handleFailed(ctx, actx, task, input, taskRead, taskWrite, prdWrite)
	}

	actx.Resolve("Run", actx.EntityID)
	return action.OK()
}

func (a *CompleteRunAction) handleCompleted(
	ctx context.Context,
	actx *action.Context,
	task *entities.Task,
	input inputs.CompleteRunInput,
	taskRead entity.ReadRepository[entities.Task],
	taskWrite entity.WriteRepository[entities.Task],
	prdWrite entity.WriteRepository[entities.PRD],
	now time.Time,
) {
	// a. Update task: done, accumulate cost (critical path).
	summary := helpers.Truncate(input.Result, 500)
	if _, err := taskWrite.Update(ctx).
		WhereID(task.ID).
		Set(fields.Task.Status.Name(), entities.TaskDone).
		Set(fields.Task.CostUSD.Name(), task.CostUSD+input.CostUSD).
		Set(fields.Task.CommitHash.Name(), input.CommitHash).
		Set(fields.Task.Summary.Name(), summary).
		Set(fields.Task.CompletedAt.Name(), now).
		Exec(ctx); err != nil {
		fmt.Printf("[factory] ERROR: failed to update task %s to done: %v\n", task.ID, err)
		return
	}

	// b. Unblock dependent tasks.
	unblockedIDs := unblockDependents(ctx, taskRead, taskWrite, task.ID)

	// c. Update PRD counters and check completion (non-critical, log and continue).
	if err := updatePRDCounters(ctx, taskRead, action.ReadRepo[entities.PRD](actx), prdWrite, task.PrdID); err != nil {
		fmt.Printf("[factory] WARN: failed to update PRD counters for %s: %v\n", task.PrdID, err)
	}

	// d. Trigger next queued task — prefer newly unblocked, else next by sequence.
	nextID := ""
	if len(unblockedIDs) > 0 {
		nextID = unblockedIDs[0]
	} else {
		nextID = findNextQueued(ctx, taskRead, task.PrdID)
	}
	if nextID != "" && a.JobClient != nil && a.WorkflowJob != nil {
		if _, err := a.JobClient.Dispatch(ctx, a.WorkflowJob, map[string]string{
			"task_id": nextID,
		}); err != nil {
			fmt.Printf("[factory] WARN: failed to dispatch workflow for task %s: %v\n", nextID, err)
		}
	}
}

func (a *CompleteRunAction) handleFailed(
	ctx context.Context,
	actx *action.Context,
	task *entities.Task,
	input inputs.CompleteRunInput,
	taskRead entity.ReadRepository[entities.Task],
	taskWrite entity.WriteRepository[entities.Task],
	prdWrite entity.WriteRepository[entities.PRD],
) {
	// a. Accumulate cost, increment run count.
	newRunCount := task.RunCount + 1
	newCost := task.CostUSD + input.CostUSD

	if newRunCount < task.MaxRetries {
		// Retry: requeue the task (critical path).
		if _, err := taskWrite.Update(ctx).
			WhereID(task.ID).
			Set(fields.Task.Status.Name(), entities.TaskQueued).
			Set(fields.Task.CostUSD.Name(), newCost).
			Set(fields.Task.RunCount.Name(), newRunCount).
			Exec(ctx); err != nil {
			fmt.Printf("[factory] ERROR: failed to requeue task %s: %v\n", task.ID, err)
			return
		}

		// Enqueue retry.
		if a.JobClient != nil && a.WorkflowJob != nil {
			if _, err := a.JobClient.Dispatch(ctx, a.WorkflowJob, map[string]string{
				"task_id": task.ID,
			}); err != nil {
				fmt.Printf("[factory] WARN: failed to dispatch retry for task %s: %v\n", task.ID, err)
			}
		}
	} else {
		// Exhausted retries: mark failed (critical path).
		if _, err := taskWrite.Update(ctx).
			WhereID(task.ID).
			Set(fields.Task.Status.Name(), entities.TaskFailed).
			Set(fields.Task.CostUSD.Name(), newCost).
			Set(fields.Task.RunCount.Name(), newRunCount).
			Exec(ctx); err != nil {
			fmt.Printf("[factory] ERROR: failed to mark task %s as failed: %v\n", task.ID, err)
			return
		}

		// Cascade failure to downstream dependents.
		cascadeFailure(ctx, taskRead, taskWrite, task.ID, task.PrdID)

		// Update PRD counters (non-critical, log and continue).
		if err := updatePRDCounters(ctx, taskRead, action.ReadRepo[entities.PRD](actx), prdWrite, task.PrdID); err != nil {
			fmt.Printf("[factory] WARN: failed to update PRD counters for %s: %v\n", task.PrdID, err)
		}
	}
}

// --- Helper functions ---

// unblockDependents finds blocked tasks depending on completedTaskID
// and transitions them to "queued" if all their deps are met.
// Returns IDs of newly unblocked tasks.
func unblockDependents(
	ctx context.Context,
	taskRead entity.ReadRepository[entities.Task],
	taskWrite entity.WriteRepository[entities.Task],
	completedTaskID string,
) []string {
	result, err := taskRead.FindMany(ctx, entity.FindOptions{
		Filters: []entity.FilterCondition{
			{Field: "status", Operator: entity.OpEq, Value: entities.TaskBlocked},
		},
	})
	if err != nil {
		return nil
	}

	var unblocked []string
	for _, t := range result.Data {
		if !helpers.ContainsDep(t.DependsOn, completedTaskID) {
			continue
		}
		if !allDepsMet(ctx, taskRead, helpers.ParseDeps(t.DependsOn)) {
			continue
		}
		_, err := taskWrite.Update(ctx).
			WhereID(t.ID).
			Set(fields.Task.Status.Name(), entities.TaskQueued).
			Exec(ctx)
		if err == nil {
			unblocked = append(unblocked, t.ID)
		}
	}
	return unblocked
}

// allDepsMet checks if every dep ID has status "done".
func allDepsMet(ctx context.Context, taskRead entity.ReadRepository[entities.Task], depIDs []string) bool {
	for _, id := range depIDs {
		t, err := taskRead.FindOne(ctx, entity.FindOneOptions{ID: id})
		if err != nil || t == nil || t.Status != entities.TaskDone {
			return false
		}
	}
	return true
}

// cascadeFailure recursively marks tasks that depend on failedTaskID as "failed".
func cascadeFailure(
	ctx context.Context,
	taskRead entity.ReadRepository[entities.Task],
	taskWrite entity.WriteRepository[entities.Task],
	failedTaskID string,
	prdID string,
) {
	// Load all non-terminal tasks in the same PRD.
	result, err := taskRead.FindMany(ctx, entity.FindOptions{
		Filters: []entity.FilterCondition{
			{Field: "prd_id", Operator: entity.OpEq, Value: prdID},
		},
	})
	if err != nil {
		return
	}

	// Find direct dependents of the failed task.
	for _, t := range result.Data {
		if t.Status == entities.TaskDone || t.Status == entities.TaskFailed || t.Status == entities.TaskCancelled {
			continue
		}
		if !helpers.ContainsDep(t.DependsOn, failedTaskID) {
			continue
		}
		_, err := taskWrite.Update(ctx).
			WhereID(t.ID).
			Set(fields.Task.Status.Name(), entities.TaskFailed).
			Exec(ctx)
		if err == nil {
			// Recurse: cascade to tasks depending on this one.
			cascadeFailure(ctx, taskRead, taskWrite, t.ID, prdID)
		}
	}
}

// findNextQueued finds the next queued task in the PRD ordered by sequence.
func findNextQueued(ctx context.Context, taskRead entity.ReadRepository[entities.Task], prdID string) string {
	result, err := taskRead.FindMany(ctx, entity.FindOptions{
		Filters: []entity.FilterCondition{
			{Field: "prd_id", Operator: entity.OpEq, Value: prdID},
			{Field: "status", Operator: entity.OpEq, Value: entities.TaskQueued},
		},
		Sort:       &entity.SortParams{Field: "sequence", Order: "asc"},
		Pagination: &entity.PaginationParams{Limit: 1},
	})
	if err != nil || len(result.Data) == 0 {
		return ""
	}
	return result.Data[0].ID
}

// updatePRDCounters recalculates completed/failed task counts and cost,
// and checks if the PRD is fully complete.
func updatePRDCounters(
	ctx context.Context,
	taskRead entity.ReadRepository[entities.Task],
	prdRead entity.ReadRepository[entities.PRD],
	prdWrite entity.WriteRepository[entities.PRD],
	prdID string,
) error {
	// Count task statuses for this PRD.
	result, err := taskRead.FindMany(ctx, entity.FindOptions{
		Filters: []entity.FilterCondition{
			{Field: "prd_id", Operator: entity.OpEq, Value: prdID},
		},
	})
	if err != nil {
		return fmt.Errorf("load tasks for PRD %s: %w", prdID, err)
	}

	var completed, failed, total int
	var totalCost float64
	for _, t := range result.Data {
		total++
		totalCost += t.CostUSD
		switch t.Status {
		case entities.TaskDone:
			completed++
		case entities.TaskFailed:
			failed++
		}
	}

	if _, err := prdWrite.Update(ctx).
		WhereID(prdID).
		Set(fields.PRD.CompletedTasks.Name(), completed).
		Set(fields.PRD.FailedTasks.Name(), failed).
		Set(fields.PRD.TotalCostUSD.Name(), totalCost).
		Exec(ctx); err != nil {
		return fmt.Errorf("update PRD counters %s: %w", prdID, err)
	}

	// Check if PRD is complete: all tasks are in a terminal state.
	if total > 0 && (completed+failed) == total {
		now := time.Now()
		status := entities.PRDCompleted
		if failed > 0 {
			status = entities.PRDFailed
		}
		if _, err := prdWrite.Update(ctx).
			WhereID(prdID).
			Set(fields.PRD.Status.Name(), status).
			Set(fields.PRD.CompletedAt.Name(), now).
			Exec(ctx); err != nil {
			return fmt.Errorf("update PRD status %s: %w", prdID, err)
		}

		eventType := events.PRDCompletedName
		if failed > 0 {
			eventType = events.PRDFailedName
		}
		service.EmitEvent(ctx, service.PendingEvent{
			EntityType: "PRD",
			EntityID:   prdID,
			Name:       eventType,
			Data: events.PRDPayload{
				PRDID:        prdID,
				TaskCount:    total,
				TotalCostUSD: totalCost,
			},
		})
	}
	return nil
}

