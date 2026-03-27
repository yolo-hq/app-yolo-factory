package actions

import (
	"context"
	"time"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/inputs"
)

type CompleteRunAction struct {
	action.TypedInput[inputs.CompleteRunInput]
}


func (a *CompleteRunAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	input := a.Input(actx)

	// Load run
	run, r2 := action.FindOrFail(ctx, action.ReadRepo[entities.Run](actx), input.RunID)
	if r2 != nil {
		return *r2
	}

	// Update run
	now := time.Now()
	action.WriteRepo[entities.Run](actx).Update(ctx).
		WhereID(input.RunID).
		Set("status", input.Status).
		Set("cost", input.Cost).
		Set("duration", input.Duration).
		Set("error", input.Error).
		Set("commit_hash", input.CommitHash).
		Set("log_url", input.LogURL).
		Set("completed_at", &now).
		Exec(ctx)

	// Load task
	task, r3 := action.FindOrFail(ctx, action.ReadRepo[entities.Task](actx), run.TaskID)
	if r3 != nil {
		return *r3
	}

	// Map run status to task status
	taskStatus := input.Status
	if taskStatus == "complete" {
		taskStatus = "done"
	}

	// If failed and retries left, re-queue
	if taskStatus == "failed" && task.RunCount < task.MaxRetries {
		taskStatus = "queued"
	}

	// Update task
	action.WriteRepo[entities.Task](actx).Update(ctx).
		WhereID(run.TaskID).
		Set("status", taskStatus).
		Set("cost", task.Cost+input.Cost).
		Exec(ctx)

	// If done, unblock dependents
	if taskStatus == "done" {
		a.unblockDependents(ctx, actx, run.TaskID)
	}

	// Set entity ID for event consumers and response
	actx.Resolve("Run", input.RunID)
	return action.OK()
}

func (a *CompleteRunAction) unblockDependents(ctx context.Context, actx *action.Context, completedTaskID string) {
	blocked, _ := action.ReadRepo[entities.Task](actx).FindMany(ctx, entity.FindOptions{
		Filters: []entity.FilterCondition{
			{Field: "status", Operator: entity.OpEq, Value: "blocked"},
		},
	})

	for _, task := range blocked.Data {
		deps := parseDeps(task.DependsOn)

		dependsOnCompleted := false
		for _, d := range deps {
			if d == completedTaskID {
				dependsOnCompleted = true
				break
			}
		}
		if !dependsOnCompleted {
			continue
		}

		if allDepsDone(ctx, action.ReadRepo[entities.Task](actx), deps) {
			action.WriteRepo[entities.Task](actx).Update(ctx).
				WhereID(task.ID).
				Set("status", "queued").
				Exec(ctx)
		}
	}
}
