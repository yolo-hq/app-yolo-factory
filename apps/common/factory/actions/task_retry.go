package actions

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// RetryTaskAction retries a failed task, resetting it to queued.
type RetryTaskAction struct {
	action.RequirePolicy[policies.CanRetryTaskPolicy]
	action.TypedInput[inputs.RetryTaskInput]
}

func (a *RetryTaskAction) Description() string { return "Retry a failed task" }

func (a *RetryTaskAction) Execute(ctx context.Context, actx *action.Context) error {
	input := a.Input(actx)

	_, err := sm.Task.Retry(ctx, actx, actx.EntityID, write.Set{
		fields.Task.Model.When(input.Model != "").Value(input.Model),
	})
	if err != nil {
		return fmt.Errorf("retry-task: %w", err)
	}
	return nil
}
