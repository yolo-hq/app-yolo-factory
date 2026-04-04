package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/inputs"
)

// RetryTaskAction retries a failed task, resetting it to queued.
type RetryTaskAction struct {
	action.TypedInput[inputs.RetryTaskInput]
}

func (a *RetryTaskAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	task, r := action.FindOrFail[entities.Task](ctx, action.ReadRepo[entities.Task](actx), actx.EntityID)
	if r != nil {
		return *r
	}

	if task.Status != "failed" {
		return action.Failure("task must be failed to retry")
	}

	input := a.Input(actx)

	set := write.Set{
		write.NewField[string]("status").Value("queued"),
	}
	if input.Model != "" {
		set = append(set, write.NewField[string]("model").Value(input.Model))
	}

	_, err := action.Write[entities.Task](actx).Exec(ctx, write.Update{
		ID:  actx.EntityID,
		Set: set,
	})
	if err != nil {
		return action.Failure(err.Error())
	}

	actx.Resolve("Task", actx.EntityID)
	return action.OK()
}
