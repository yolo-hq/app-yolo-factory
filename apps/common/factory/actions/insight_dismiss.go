package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
)

// InsightDismiss dismisses an insight.
//
// @policy CanDismissInsightPolicy
func InsightDismiss(ctx context.Context, actx *action.Context) error {
	_, err := sm.Insight.Dismiss(ctx, actx, actx.EntityID, nil)
	if errors.Is(err, action.ErrStaleState) {
		return action.Fail("insight is already applied or dismissed")
	}
	if err != nil {
		return fmt.Errorf("dismiss-insight: %w", err)
	}
	return nil
}
