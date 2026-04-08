package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// RetryTaskAction retries a failed task, resetting it to queued.
type RetryTaskAction struct {
	action.TypedInput[inputs.RetryTaskInput]
	action.RequirePolicy[policies.CanRetryTaskPolicy]
}

func (a *RetryTaskAction) Description() string { return "Retry a failed task" }

func (a *RetryTaskAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	input := a.Input(actx)

	set := write.Set{
		write.NewField[string]("status").Value(entities.TaskQueued),
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
