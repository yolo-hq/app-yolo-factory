package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

// TaskRetry retries a failed task, resetting it to queued.
//
// @policy CanRetryTaskPolicy
func TaskRetry(ctx context.Context, actx *action.Context, in inputs.RetryTaskInput) error {
	_, err := sm.Task.Retry(ctx, actx, actx.EntityID, write.Set{
		fields.Task.Model.When(in.Model != "").Value(in.Model),
	})
	if errors.Is(err, action.ErrStaleState) {
		return action.Fail("task is not in failed state")
	}
	if err != nil {
		return fmt.Errorf("retry-task: %w", err)
	}
	return nil
}
