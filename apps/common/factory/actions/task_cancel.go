package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// CancelTaskAction cancels a task.
type CancelTaskAction struct {
	action.NoInput
	action.RequirePolicy[policies.CanCancelTaskPolicy]
}

func (a *CancelTaskAction) Description() string { return "Cancel a non-terminal task" }

func (a *CancelTaskAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	if r := action.ExecUpdate[entities.Task](ctx, actx, write.Set{
		write.NewField[string]("status").Value(entities.TaskCancelled),
	}); r != nil {
		return *r
	}
	actx.Resolve("Task", actx.EntityID)
	return action.OK()
}
