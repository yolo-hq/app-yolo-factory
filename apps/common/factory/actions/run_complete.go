package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/jobs"
	"github.com/yolo-hq/yolo/core/write"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/.yolo/svc"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/events"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/helpers"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
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

			// Aggregations (replace updatePRDCounters helper).
			CompletedCount int     `field:"tasks" aggregate:"count" filter:"status:done"`
			FailedCount    int     `field:"tasks" aggregate:"count" filter:"status:failed"`
			TotalCost      float64 `field:"tasks" aggregate:"sum:cost_usd"`
			TotalCount     int     `field:"tasks" aggregate:"count"`

			// Has-many slices (replace unblockDependents + findNextQueued helpers).
			BlockedTasks []TaskRef `field:"tasks" filter:"status:blocked"`
			QueuedTasks  []TaskRef `field:"tasks" filter:"status:queued" limit:"1"`
		} `field:"prd"`
	} `field:"task"`
}

// CompleteRunAction records run completion and drives the task/PRD state machine.
type CompleteRunAction struct {
	action.TypedInput[inputs.CompleteRunInput]
	action.PublicAccess
	action.TypedData[CompleteRunData]
}

func (a *CompleteRunAction) Description() string {
	return "Record run completion and advance state machine"
}

func (a *CompleteRunAction) Execute(ctx context.Context, actx *action.Context) error {
	data := a.Data(actx)
	input := a.Input(actx)
	now := time.Now()

	// 1. Update the run with completion data.
	if _, err := action.Write[entities.Run](actx).Exec(ctx, write.Update{
		ID:        actx.EntityID,
		FromInput: input,
		Set:       write.Set{fields.Run.CompletedAt.Value(&now)},
	}); err != nil {
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
	if _, err := action.Write[entities.Task](actx).Exec(ctx, write.Update{
		ID: data.Task.ID,
		Set: write.Set{
			fields.Task.Status.Value(string(enums.TaskStatusDone)),
			fields.Task.CostUSD.Incr(input.CostUSD),
			fields.Task.CommitHash.Value(input.CommitHash),
			fields.Task.Summary.Value(helpers.Truncate(input.Result, 500)),
			fields.Task.CompletedAt.Value(&now),
		},
	}); err != nil {
		fmt.Printf("[factory] ERROR: failed to update task %s to done: %v\n", data.Task.ID, err)
		return
	}

	// b. Unblock dependents — build list then issue a single UpdateMany.
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
		if _, err := action.Write[entities.Task](actx).Exec(ctx, write.UpdateMany{
			IDs: toUnblock,
			Set: write.Set{fields.Task.Status.Value(string(enums.TaskStatusQueued))},
		}); err != nil {
			fmt.Printf("[factory] WARN: failed to unblock tasks: %v\n", err)
		}
	}

	// c. Update PRD counters from pre-computed aggregations.
	newCompleted := data.Task.PRD.CompletedCount + 1
	newCost := data.Task.PRD.TotalCost + input.CostUSD
	if _, err := action.Write[entities.PRD](actx).Exec(ctx, write.Update{
		ID: data.Task.PRD.ID,
		Set: write.Set{
			fields.PRD.CompletedTasks.Value(newCompleted),
			fields.PRD.TotalCostUSD.Value(newCost),
		},
	}); err != nil {
		fmt.Printf("[factory] WARN: failed to update PRD counters for %s: %v\n", data.Task.PRD.ID, err)
	}

	// d. Check PRD completion.
	if data.Task.PRD.TotalCount > 0 && (newCompleted+data.Task.PRD.FailedCount) == data.Task.PRD.TotalCount {
		status := string(enums.PRDStatusCompleted)
		if data.Task.PRD.FailedCount > 0 {
			status = string(enums.PRDStatusFailed)
		}
		if _, err := action.Write[entities.PRD](actx).Exec(ctx, write.Update{
			ID: data.Task.PRD.ID,
			Set: write.Set{
				fields.PRD.Status.Value(status),
				fields.PRD.CompletedAt.Value(&now),
			},
		}); err != nil {
			fmt.Printf("[factory] WARN: failed to update PRD status %s: %v\n", data.Task.PRD.ID, err)
		} else if data.Task.PRD.FailedCount > 0 {
			events.PRDFailed.Emit(ctx, data.Task.PRD.ID)
		} else {
			events.PRDCompleted.Emit(ctx, data.Task.PRD.ID)
		}
	}

	// e. Dispatch next task — prefer newly unblocked, else pre-loaded queued.
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

func (a *CompleteRunAction) handleFailed(
	ctx context.Context,
	actx *action.Context,
	data CompleteRunData,
	input inputs.CompleteRunInput,
) {
	newRunCount := data.Task.RunCount + 1

	if newRunCount < data.Task.MaxRetries {
		// Retry: requeue task with atomic cost/run_count increment.
		if _, err := action.Write[entities.Task](actx).Exec(ctx, write.Update{
			ID: data.Task.ID,
			Set: write.Set{
				fields.Task.Status.Value(string(enums.TaskStatusQueued)),
				fields.Task.CostUSD.Incr(input.CostUSD),
				fields.Task.RunCount.Incr(1),
			},
		}); err != nil {
			fmt.Printf("[factory] ERROR: failed to requeue task %s: %v\n", data.Task.ID, err)
			return
		}
		jobs.Defer(ctx, &factoryjobs.ExecuteWorkflowJob{TaskID: data.Task.ID})
		return
	}

	// Exhausted retries: mark failed.
	if _, err := action.Write[entities.Task](actx).Exec(ctx, write.Update{
		ID: data.Task.ID,
		Set: write.Set{
			fields.Task.Status.Value(string(enums.TaskStatusFailed)),
			fields.Task.CostUSD.Incr(input.CostUSD),
			fields.Task.RunCount.Incr(1),
		},
	}); err != nil {
		fmt.Printf("[factory] ERROR: failed to mark task %s as failed: %v\n", data.Task.ID, err)
		return
	}

	// Cascade failure via the completion service (recursive graph walk).
	if svc.S.RunCompletion != nil {
		if err := svc.S.RunCompletion.CascadeFailure(ctx, data.Task.ID, data.Task.PRD.ID); err != nil {
			fmt.Printf("[factory] WARN: cascade failed for %s: %v\n", data.Task.ID, err)
		}
	}

	// Update PRD counters using pre-computed aggregations.
	newFailed := data.Task.PRD.FailedCount + 1
	newCost := data.Task.PRD.TotalCost + input.CostUSD
	if _, err := action.Write[entities.PRD](actx).Exec(ctx, write.Update{
		ID: data.Task.PRD.ID,
		Set: write.Set{
			fields.PRD.FailedTasks.Value(newFailed),
			fields.PRD.TotalCostUSD.Value(newCost),
		},
	}); err != nil {
		fmt.Printf("[factory] WARN: failed to update PRD counters for %s: %v\n", data.Task.PRD.ID, err)
	}
}
