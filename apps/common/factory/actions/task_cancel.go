package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
)

// TaskCancel cancels a non-terminal task.
//
// @policy CanCancelTaskPolicy
func TaskCancel(ctx context.Context, actx *action.Context) error {
	_, err := sm.Task.Cancel(ctx, actx, actx.EntityID, nil)
	if errors.Is(err, action.ErrStaleState) {
		return action.Fail("task already in a terminal state")
	}
	if err != nil {
		return fmt.Errorf("cancel-task: %w", err)
	}
	return nil
}
