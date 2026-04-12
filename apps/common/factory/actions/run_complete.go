package actions

import (
	"context"
	"log/slog"
	"time"

	yolostrings "github.com/yolo-hq/yolo/core/strings"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/jobs"
	"github.com/yolo-hq/yolo/core/projection"
	"github.com/yolo-hq/yolo/core/write"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/.yolo/repos"
	"github.com/yolo-hq/app-yolo-factory/.yolo/svc"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/events"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/helpers"
	factoryjobs "github.com/yolo-hq/app-yolo-factory/apps/common/factory/jobs"
)

// TaskRef is a minimal reference to a task used inside CompleteRunData slices.
type TaskRef struct {
	ID        string `field:"id"`
	DependsOn string `field:"depends_on"`
}

// CompleteRunData declares the Run entity fields and relation data this action
// reads. The framework resolves the belongs_to chain Run → Task → PRD and
// pre-computes the aggregations + filtered has_many slices, replacing all the
// manual loading helpers this action used to carry.
type CompleteRunData struct {
	projection.For[entities.Run]

	Status string `field:"status"`
	Task   struct {
		ID         string  `field:"id"`
		Status     string  `field:"status"`
		CostUSD    float64 `field:"cost_usd"`
		RunCount   int     `field:"run_count"`
		MaxRetries int     `field:"max_retries"`
		DependsOn  string  `field:"depends_on"`
		PrdID      string  `field:"prd_id"`
		PRD        struct {
			ID string `field:"id"`

			// Aggregations for terminal-check logic.
			CompletedCount int `field:"tasks" aggregate:"count" filter:"status:done"`
			FailedCount    int `field:"tasks" aggregate:"count" filter:"status:failed"`
			TotalCount     int `field:"tasks" aggregate:"count"`

			// Has-many slices (replace unblockDependents + findNextQueued helpers).
			BlockedTasks []TaskRef `field:"tasks" filter:"status:blocked"`
			QueuedTasks  []TaskRef `field:"tasks" filter:"status:queued" limit:"1"`
		} `field:"prd"`
	} `field:"task"`
}

// CompleteRunAction records run completion and drives the task/PRD state machine.
type CompleteRunAction struct {
	action.SkipAllPolicies
	action.TypedInput[inputs.CompleteRunInput]
	action.Projection[CompleteRunData]
}

func (a *CompleteRunAction) Description() string {
	return "Record run completion and advance state machine"
}

func (a *CompleteRunAction) Execute(ctx context.Context, actx *action.Context) error {
	data := a.Data(actx)
	input := a.Input(actx)
	now := time.Now()

	// 1. Update the run with completion data.
	if _, err := repos.Run.UpdateFromInput(ctx, actx, actx.EntityID, input,
		fields.Run.CompletedAt.Value(&now),
	); err != nil {
		return err
	}

	// 2. Branch on run outcome.
	switch input.Status {
	case string(enums.RunStatusCompleted):
		a.handleCompleted(ctx, actx, data, input, now)
	case string(enums.RunStatusFailed):
		a.handleFailed(ctx, actx, data, input)
	}

	return nil
}

func (a *CompleteRunAction) handleCompleted(
	ctx context.Context,
	actx *action.Context,
	data CompleteRunData,
	input inputs.CompleteRunInput,
	now time.Time,
) {
	// a. Mark task done, accumulate cost atomically.
	if _, err := repos.Task.Update(ctx, actx, data.Task.ID, write.Set{
		fields.Task.Status.Value(string(enums.TaskStatusDone)),
		fields.Task.CostUSD.Incr(input.CostUSD),
		fields.Task.CommitHash.Value(input.CommitHash),
		fields.Task.Summary.Value(yolostrings.Truncate(input.Result, 500)),
		fields.Task.CompletedAt.Value(&now),
	}); err != nil {
		slog.Error("failed to update task to done", "task_id", data.Task.ID, "error", err)
		return
	}

	// b. Unblock dependents and advance PRD state.
	toUnblock := a.unblockDependents(ctx, actx, data)
	a.advancePRD(ctx, actx, data, now)

	// c. Dispatch next task — prefer newly unblocked, else pre-loaded queued.
	nextID := ""
	if len(toUnblock) > 0 {
		nextID = toUnblock[0]
	} else if len(data.Task.PRD.QueuedTasks) > 0 {
		nextID = data.Task.PRD.QueuedTasks[0].ID
	}
	if nextID != "" {
		jobs.Defer(ctx, &factoryjobs.ExecuteWorkflowJob{TaskID: nextID})
	}
}

// unblockDependents transitions blocked tasks whose dependencies are all met to
// "queued" and returns the IDs of tasks that were unblocked.
func (a *CompleteRunAction) unblockDependents(
	ctx context.Context,
	actx *action.Context,
	data CompleteRunData,
) []string {
	var toUnblock []string
	for _, blocked := range data.Task.PRD.BlockedTasks {
		if !helpers.ContainsDep(blocked.DependsOn, data.Task.ID) {
			continue
		}
		if svc.S.RunCompletion != nil {
			allMet, err := svc.S.RunCompletion.AllDepsMet(ctx, helpers.ParseDeps(blocked.DependsOn))
			if err != nil || !allMet {
				continue
			}
		}
		toUnblock = append(toUnblock, blocked.ID)
	}
	if len(toUnblock) > 0 {
		if _, err := repos.Task.UpdateMany(ctx, actx, toUnblock, write.Set{
			fields.Task.Status.Value(string(enums.TaskStatusQueued)),
		}); err != nil {
			slog.Warn("failed to unblock tasks", "error", err)
		}
	}
	return toUnblock
}

// advancePRD checks whether all tasks are terminal and transitions the PRD
// to completed or failed, emitting the appropriate event.
// CompletedTasks and TotalCostUSD are now virtual computed fields — no manual
// counter updates needed.
func (a *CompleteRunAction) advancePRD(
	ctx context.Context,
	actx *action.Context,
	data CompleteRunData,
	now time.Time,
) {
	newCompleted := data.Task.PRD.CompletedCount + 1
	if data.Task.PRD.TotalCount == 0 || (newCompleted+data.Task.PRD.FailedCount) != data.Task.PRD.TotalCount {
		return
	}

	status := string(enums.PRDStatusCompleted)
	if data.Task.PRD.FailedCount > 0 {
		status = string(enums.PRDStatusFailed)
	}
	if _, err := repos.PRD.Update(ctx, actx, data.Task.PRD.ID, write.Set{
		fields.PRD.Status.Value(status),
		fields.PRD.CompletedAt.Value(&now),
	}); err != nil {
		slog.Warn("failed to update PRD status", "prd_id", data.Task.PRD.ID, "error", err)
		return
	}
	if data.Task.PRD.FailedCount > 0 {
		events.PRDFailed.Emit(ctx, data.Task.PRD.ID)
	} else {
		events.PRDCompleted.Emit(ctx, data.Task.PRD.ID)
	}
}

func (a *CompleteRunAction) handleFailed(
	ctx context.Context,
	actx *action.Context,
	data CompleteRunData,
	input inputs.CompleteRunInput,
) {
	newRunCount := data.Task.RunCount + 1

	if newRunCount < data.Task.MaxRetries {
		// Retry: requeue task with atomic cost/run_count increment.
		if _, err := repos.Task.Update(ctx, actx, data.Task.ID, write.Set{
			fields.Task.Status.Value(string(enums.TaskStatusQueued)),
			fields.Task.CostUSD.Incr(input.CostUSD),
			fields.Task.RunCount.Incr(1),
		}); err != nil {
			slog.Error("failed to requeue task", "task_id", data.Task.ID, "error", err)
			return
		}
		jobs.Defer(ctx, &factoryjobs.ExecuteWorkflowJob{TaskID: data.Task.ID})
		return
	}

	// Exhausted retries: mark failed.
	if _, err := repos.Task.Update(ctx, actx, data.Task.ID, write.Set{
		fields.Task.Status.Value(string(enums.TaskStatusFailed)),
		fields.Task.CostUSD.Incr(input.CostUSD),
		fields.Task.RunCount.Incr(1),
	}); err != nil {
		slog.Error("failed to mark task as failed", "task_id", data.Task.ID, "error", err)
		return
	}

	// Cascade failure via the completion service (recursive graph walk).
	if svc.S.RunCompletion != nil {
		if err := svc.S.RunCompletion.CascadeFailure(ctx, data.Task.ID, data.Task.PRD.ID); err != nil {
			slog.Warn("cascade failure failed", "task_id", data.Task.ID, "error", err)
		}
	}

	// Update PRD failed counter. TotalCostUSD is now a virtual computed field.
	newFailed := data.Task.PRD.FailedCount + 1
	if _, err := repos.PRD.Update(ctx, actx, data.Task.PRD.ID, write.Set{
		fields.PRD.FailedTasks.Value(newFailed),
	}); err != nil {
		slog.Warn("failed to update PRD counters", "prd_id", data.Task.PRD.ID, "error", err)
	}
}
