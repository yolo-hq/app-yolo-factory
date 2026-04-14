package actions

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/jobs"
	"github.com/yolo-hq/yolo/core/projection"
	"github.com/yolo-hq/yolo/core/read"
	yolostrings "github.com/yolo-hq/yolo/core/strings"
	"github.com/yolo-hq/yolo/core/write"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/.yolo/repos"
	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
	"github.com/yolo-hq/app-yolo-factory/.yolo/svc"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/helpers"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
	factoryjobs "github.com/yolo-hq/app-yolo-factory/apps/common/factory/jobs"
)

// TaskRef is a minimal task projection for dependency traversal.
type TaskRef struct {
	projection.For[entities.Task]
	ID        string `field:"id"`
	DependsOn string `field:"depends_on"`
}

// CompleteRunData declares the Run entity fields and relation data this action reads.
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

	// 1. Transition the run via SM, carrying input metadata as extras.
	//    We split per outcome so the SM enforces running → {completed,failed}.
	runExtras := write.Set{
		fields.Run.CompletedAt.Value(&now),
		fields.Run.CostUSD.When(input.CostUSD != 0).Value(input.CostUSD),
		fields.Run.TokensIn.When(input.TokensIn != 0).Value(input.TokensIn),
		fields.Run.TokensOut.When(input.TokensOut != 0).Value(input.TokensOut),
		fields.Run.DurationMs.When(input.DurationMS != 0).Value(input.DurationMS),
		fields.Run.NumTurns.When(input.NumTurns != 0).Value(input.NumTurns),
		fields.Run.Error.When(input.Error != "").Value(input.Error),
		fields.Run.CommitHash.When(input.CommitHash != "").Value(input.CommitHash),
		fields.Run.FilesChanged.When(input.FilesChanged != "").Value(input.FilesChanged),
		fields.Run.Result.When(input.Result != "").Value(input.Result),
		fields.Run.SessionID.When(input.SessionID != "").Value(input.SessionID),
	}

	var runErr error
	switch input.Status {
	case string(enums.RunStatusCompleted):
		_, runErr = sm.Run.Complete(ctx, actx, actx.EntityID, runExtras)
	case string(enums.RunStatusFailed):
		_, runErr = sm.Run.Fail(ctx, actx, actx.EntityID, runExtras)
	default:
		// Unknown / unsupported terminal status — fall back to direct update so
		// callers passing legacy values still persist their data.
		_, runErr = repos.Run.UpdateFromInput(ctx, actx, actx.EntityID, input,
			fields.Run.CompletedAt.Value(&now),
		)
	}
	if errors.Is(runErr, action.ErrStaleState) {
		return action.Fail("run already in a terminal state")
	}
	if runErr != nil {
		return runErr
	}

	// 2. Branch on run outcome.
	switch input.Status {
	case string(enums.RunStatusCompleted):
		handleCompleted(ctx, actx, data, input, now)
	case string(enums.RunStatusFailed):
		handleFailed(ctx, actx, data, input)
	}

	return nil
}

func handleCompleted(
	ctx context.Context,
	actx *action.Context,
	data CompleteRunData,
	input inputs.CompleteRunInput,
	now time.Time,
) {
	// a. Mark task done via SM, accumulate cost atomically.
	if _, err := sm.Task.Complete(ctx, actx, data.Task.ID, write.Set{
		fields.Task.CostUSD.Incr(input.CostUSD),
		fields.Task.CommitHash.Value(input.CommitHash),
		fields.Task.Summary.Value(yolostrings.Truncate(input.Result, 500)),
		fields.Task.CompletedAt.Value(&now),
	}); err != nil {
		if errors.Is(err, action.ErrStaleState) {
			slog.Warn("task already in terminal state on completion", "task_id", data.Task.ID)
			return
		}
		slog.Error("failed to update task to done", "task_id", data.Task.ID, "error", err)
		return
	}

	// b. Unblock dependents and advance PRD state.
	toUnblock := unblockDependents(ctx, actx, data)
	advancePRD(ctx, actx, data, now)

	// c. Dispatch next task — prefer newly unblocked, else find next queued.
	nextID := ""
	if len(toUnblock) > 0 {
		nextID = toUnblock[0]
	} else {
		queued, err := read.FindMany[TaskRef](ctx,
			read.Eq(fields.Task.PrdID.Name(), data.Task.PrdID),
			read.Eq(fields.Task.Status.Name(), string(enums.TaskStatusQueued)),
			read.Limit(1),
		)
		if err == nil && len(queued) > 0 {
			nextID = queued[0].ID
		}
	}
	if nextID != "" {
		jobs.Defer(ctx, &factoryjobs.ExecuteWorkflowJob{TaskID: nextID})
	}
}

// unblockDependents transitions blocked tasks whose dependencies are all met to
// "queued" and returns the IDs of tasks that were unblocked.
func unblockDependents(
	ctx context.Context,
	actx *action.Context,
	data CompleteRunData,
) []string {
	blockedTasks, err := read.FindMany[TaskRef](ctx,
		read.Eq(fields.Task.PrdID.Name(), data.Task.PrdID),
		read.Eq(fields.Task.Status.Name(), string(enums.TaskStatusBlocked)),
	)
	if err != nil {
		slog.Warn("failed to load blocked tasks", "prd_id", data.Task.PrdID, "error", err)
		return nil
	}

	var toUnblock []string
	for _, blocked := range blockedTasks {
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
	for _, id := range toUnblock {
		if _, err := sm.Task.Unblock(ctx, actx, id, nil); err != nil && !errors.Is(err, action.ErrStaleState) {
			slog.Warn("failed to unblock task", "task_id", id, "error", err)
		}
	}
	return toUnblock
}

// advancePRD checks whether all tasks are terminal and transitions the PRD
// to completed or failed, emitting the appropriate event.
// CompletedTasks and TotalCostUSD are now virtual computed fields — no manual
// counter updates needed.
func advancePRD(
	ctx context.Context,
	actx *action.Context,
	data CompleteRunData,
	now time.Time,
) {
	newCompleted := data.Task.PRD.CompletedCount + 1
	if data.Task.PRD.TotalCount == 0 || (newCompleted+data.Task.PRD.FailedCount) != data.Task.PRD.TotalCount {
		return
	}

	extras := write.Set{fields.PRD.CompletedAt.Value(&now)}
	var err error
	if data.Task.PRD.FailedCount > 0 {
		_, err = sm.PRD.Fail(ctx, actx, data.Task.PRD.ID, extras)
	} else {
		_, err = sm.PRD.Complete(ctx, actx, data.Task.PRD.ID, extras)
	}
	if err != nil && !errors.Is(err, action.ErrStaleState) {
		slog.Warn("failed to update PRD status", "prd_id", data.Task.PRD.ID, "error", err)
	}
}

func handleFailed(
	ctx context.Context,
	actx *action.Context,
	data CompleteRunData,
	input inputs.CompleteRunInput,
) {
	newRunCount := data.Task.RunCount + 1

	if newRunCount < data.Task.MaxRetries {
		// Retry: requeue task via SM with atomic cost/run_count increment.
		if _, err := sm.Task.Requeue(ctx, actx, data.Task.ID, write.Set{
			fields.Task.CostUSD.Incr(input.CostUSD),
			fields.Task.RunCount.Incr(1),
		}); err != nil {
			if errors.Is(err, action.ErrStaleState) {
				slog.Warn("task not in running state on requeue", "task_id", data.Task.ID)
				return
			}
			slog.Error("failed to requeue task", "task_id", data.Task.ID, "error", err)
			return
		}
		jobs.Defer(ctx, &factoryjobs.ExecuteWorkflowJob{TaskID: data.Task.ID})
		return
	}

	// Exhausted retries: mark failed via SM.
	if _, err := sm.Task.Fail(ctx, actx, data.Task.ID, write.Set{
		fields.Task.CostUSD.Incr(input.CostUSD),
		fields.Task.RunCount.Incr(1),
	}); err != nil {
		if errors.Is(err, action.ErrStaleState) {
			slog.Warn("task already terminal on fail", "task_id", data.Task.ID)
			return
		}
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
