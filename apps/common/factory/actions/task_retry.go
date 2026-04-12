package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
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

func (a *RetryTaskAction) Execute(ctx context.Context, actx *action.Context) error {
	input := a.Input(actx)

	_, err := action.Write[entities.Task](actx).Exec(ctx, write.Update{
		ID: actx.EntityID,
		Set: write.Set{
			fields.Task.Status.Value(string(enums.TaskStatusQueued)),
			fields.Task.Model.When(input.Model != "").Value(input.Model),
		},
	})
	return err
}
