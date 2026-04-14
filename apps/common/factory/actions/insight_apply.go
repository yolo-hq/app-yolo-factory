package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
)

// InsightApply applies an acknowledged insight.
//
// @policy CanApplyInsightPolicy
func InsightApply(ctx context.Context, actx *action.Context) error {
	_, err := sm.Insight.Apply(ctx, actx, actx.EntityID, nil)
	if errors.Is(err, action.ErrStaleState) {
		return action.Fail("insight is not acknowledged")
	}
	if err != nil {
		return fmt.Errorf("apply-insight: %w", err)
	}
	return nil
}
