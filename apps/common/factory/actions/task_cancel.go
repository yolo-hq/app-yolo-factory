package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// CancelTaskAction cancels a task.
type CancelTaskAction struct {
	action.NoInput
	action.RequirePolicy[policies.CanCancelTaskPolicy]
}

func (a *CancelTaskAction) Description() string { return "Cancel a non-terminal task" }

func (a *CancelTaskAction) Execute(ctx context.Context, actx *action.Context) error {
	_, err := action.Write[entities.Task](actx).Exec(ctx, write.Update{
		ID:  actx.EntityID,
		Set: write.Set{fields.Task.Status.Value(string(enums.TaskStatusCancelled))},
	})
	return err
}
