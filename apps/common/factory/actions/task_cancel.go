package actions

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// CancelTaskAction cancels a task.
type CancelTaskAction struct {
	action.RequirePolicy[policies.CanCancelTaskPolicy]
	action.NoInput
}

func (a *CancelTaskAction) Description() string { return "Cancel a non-terminal task" }

func (a *CancelTaskAction) Execute(ctx context.Context, actx *action.Context) error {
	_, err := sm.Task.Cancel(ctx, actx, actx.EntityID, nil)
	if err != nil {
		return fmt.Errorf("cancel-task: %w", err)
	}
	return nil
}
