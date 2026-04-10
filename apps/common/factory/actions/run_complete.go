package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/jobs"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/.yolo/svc"
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

func (a *CompleteRunAction) Description() string {
	return "Record run completion and advance state machine"
}

func (a *CompleteRunAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	run, r := action.FindOrFail[entities.Run](ctx, action.ReadRepo[entities.Run](actx), actx.EntityID)
	if r != nil {
		return *r
	}

	input := a.Input(actx)
	now := time.Now()

	// 1. Update the run with completion data.
	if _, err := action.Write[entities.Run](actx).Exec(ctx, write.Update{
		ID:        actx.EntityID,
		FromInput: input,
		Set:       write.Set{fields.Run.CompletedAt.Value(&now)},
	}); err != nil {
		return action.Failure(err.Error())
	}

	// 2. Load the parent task.
	task, err := action.ReadRepo[entities.Task](actx).FindOne(ctx, entity.FindOneOptions{ID: run.TaskID})
	if err != nil || task == nil {
		return action.Failure("failed to load task for run")
	}

	// 3. Dispatch based on run outcome.
	switch input.Status {
	case entities.RunCompleted:
		a.handleCompleted(ctx, actx, task, input, now)
	case entities.RunFailed:
		a.handleFailed(ctx, actx, task, input)
	}

	actx.Resolve("Run", actx.EntityID)
	return action.OK()
}

func (a *CompleteRunAction) handleCompleted(
	ctx context.Context,
	actx *action.Context,
	task *entities.Task,
	input inputs.CompleteRunInput,
	now time.Time,
) {
	// a. Mark task done, accumulate cost atomically.
	summary := helpers.Truncate(input.Result, 500)
	if _, err := action.Write[entities.Task](actx).Exec(ctx, write.Update{
		ID: task.ID,
		Set: write.Set{
			fields.Task.Status.Value(entities.TaskDone),
			fields.Task.CostUSD.Incr(input.CostUSD),
			fields.Task.CommitHash.Value(input.CommitHash),
			fields.Task.Summary.Value(summary),
			fields.Task.CompletedAt.Value(&now),
		},
	}); err != nil {
		fmt.Printf("[factory] ERROR: failed to update task %s to done: %v\n", task.ID, err)
		return
	}

	// b. Unblock dependent tasks.
	unblockedIDs := a.unblockDependents(ctx, actx, task.ID)

	// c. Update PRD counters and check completion (non-critical).
	if err := a.updatePRDCounters(ctx, actx, task.PrdID); err != nil {
		fmt.Printf("[factory] WARN: failed to update PRD counters for %s: %v\n", task.PrdID, err)
	}

	// d. Trigger next queued task — prefer newly unblocked, else next by sequence.
	nextID := ""
	if len(unblockedIDs) > 0 {
		nextID = unblockedIDs[0]
	} else {
		nextID = a.findNextQueued(ctx, actx, task.PrdID)
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
) {
	newRunCount := task.RunCount + 1

	if newRunCount < task.MaxRetries {
		// Retry: requeue task with atomic cost/run_count increment.
		if _, err := action.Write[entities.Task](actx).Exec(ctx, write.Update{
			ID: task.ID,
			Set: write.Set{
				fields.Task.Status.Value(entities.TaskQueued),
				fields.Task.CostUSD.Incr(input.CostUSD),
				fields.Task.RunCount.Incr(1),
			},
		}); err != nil {
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
		return
	}

	// Exhausted retries: mark failed.
	if _, err := action.Write[entities.Task](actx).Exec(ctx, write.Update{
		ID: task.ID,
		Set: write.Set{
			fields.Task.Status.Value(entities.TaskFailed),
			fields.Task.CostUSD.Incr(input.CostUSD),
			fields.Task.RunCount.Incr(1),
		},
	}); err != nil {
		fmt.Printf("[factory] ERROR: failed to mark task %s as failed: %v\n", task.ID, err)
		return
	}

	// Cascade failure via the completion service (recursive graph walk).
	if svc.S.RunCompletion != nil {
		if err := svc.S.RunCompletion.CascadeFailure(ctx, task.ID, task.PrdID); err != nil {
			fmt.Printf("[factory] WARN: cascade failed for %s: %v\n", task.ID, err)
		}
	}

	// Update PRD counters (non-critical).
	if err := a.updatePRDCounters(ctx, actx, task.PrdID); err != nil {
		fmt.Printf("[factory] WARN: failed to update PRD counters for %s: %v\n", task.PrdID, err)
	}
}

// unblockDependents finds blocked tasks depending on completedTaskID and
// transitions them to "queued" when all their deps are satisfied. Uses the
// RunCompletion service for the graph check.
func (a *CompleteRunAction) unblockDependents(ctx context.Context, actx *action.Context, completedTaskID string) []string {
	taskRead := action.ReadRepo[entities.Task](actx)
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
		if svc.S.RunCompletion != nil {
			ok, err := svc.S.RunCompletion.AllDepsMet(ctx, helpers.ParseDeps(t.DependsOn))
			if err != nil || !ok {
				continue
			}
		}
		if _, err := action.Write[entities.Task](actx).Exec(ctx, write.Update{
			ID:  t.ID,
			Set: write.Set{fields.Task.Status.Value(entities.TaskQueued)},
		}); err == nil {
			unblocked = append(unblocked, t.ID)
		}
	}
	return unblocked
}

// findNextQueued returns the next queued task ID in a PRD ordered by sequence.
func (a *CompleteRunAction) findNextQueued(ctx context.Context, actx *action.Context, prdID string) string {
	result, err := action.ReadRepo[entities.Task](actx).FindMany(ctx, entity.FindOptions{
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

// updatePRDCounters recalculates completed/failed counts and total cost from
// tasks in the PRD, writes them back, and marks the PRD complete when all
// tasks reach a terminal state.
func (a *CompleteRunAction) updatePRDCounters(ctx context.Context, actx *action.Context, prdID string) error {
	result, err := action.ReadRepo[entities.Task](actx).FindMany(ctx, entity.FindOptions{
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

	if _, err := action.Write[entities.PRD](actx).Exec(ctx, write.Update{
		ID: prdID,
		Set: write.Set{
			fields.PRD.CompletedTasks.Value(completed),
			fields.PRD.FailedTasks.Value(failed),
			fields.PRD.TotalCostUSD.Value(totalCost),
		},
	}); err != nil {
		return fmt.Errorf("update PRD counters %s: %w", prdID, err)
	}

	// Check PRD completion: all tasks in a terminal state.
	if total > 0 && (completed+failed) == total {
		now := time.Now()
		status := entities.PRDCompleted
		if failed > 0 {
			status = entities.PRDFailed
		}
		if _, err := action.Write[entities.PRD](actx).Exec(ctx, write.Update{
			ID: prdID,
			Set: write.Set{
				fields.PRD.Status.Value(status),
				fields.PRD.CompletedAt.Value(&now),
			},
		}); err != nil {
			return fmt.Errorf("update PRD status %s: %w", prdID, err)
		}
		if failed > 0 {
			events.PRDFailed.Emit(ctx, prdID)
		} else {
			events.PRDCompleted.Emit(ctx, prdID)
		}
	}
	return nil
}
