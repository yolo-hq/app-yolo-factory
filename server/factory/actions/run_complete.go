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
	RunRead   entity.ReadRepository[entities.Run]
	RunWrite  entity.WriteRepository[entities.Run]
	TaskRead  entity.ReadRepository[entities.Task]
	TaskWrite entity.WriteRepository[entities.Task]
}


func (a *CompleteRunAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	input, r := a.Input(actx)
	if r != nil {
		return *r
	}

	// Load run
	run, r2 := action.FindOrFail(ctx, a.RunRead, input.RunID)
	if r2 != nil {
		return *r2
	}

	// Update run
	now := time.Now()
	a.RunWrite.Update(ctx).
		Where(entity.FilterCondition{Field: "id", Operator: entity.OpEq, Value: input.RunID}).
		Set("status", input.Status).
		Set("cost", input.Cost).
		Set("duration", input.Duration).
		Set("error", input.Error).
		Set("commit_hash", input.CommitHash).
		Set("log_url", input.LogURL).
		Set("completed_at", &now).
		Exec(ctx)

	// Load task
	task, r3 := action.FindOrFail(ctx, a.TaskRead, run.TaskID)
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
	a.TaskWrite.Update(ctx).
		Where(entity.FilterCondition{Field: "id", Operator: entity.OpEq, Value: run.TaskID}).
		Set("status", taskStatus).
		Set("cost", task.Cost+input.Cost).
		Exec(ctx)

	// If done, unblock dependents
	if taskStatus == "done" {
		a.unblockDependents(ctx, run.TaskID)
	}

	// Set entity ID for event consumers and response
	actx.Resolve("Run", input.RunID)
	return action.OK()
}

func (a *CompleteRunAction) unblockDependents(ctx context.Context, completedTaskID string) {
	blocked, _ := a.TaskRead.FindMany(ctx, entity.FindOptions{
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

		if allDepsDone(ctx, a.TaskRead, deps) {
			a.TaskWrite.Update(ctx).
				Where(entity.FilterCondition{Field: "id", Operator: entity.OpEq, Value: task.ID}).
				Set("status", "queued").
				Exec(ctx)
		}
	}
}
