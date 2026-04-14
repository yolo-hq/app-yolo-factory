package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
)

// InsightAcknowledge acknowledges a pending insight.
//
// @policy CanAcknowledgeInsightPolicy
func InsightAcknowledge(ctx context.Context, actx *action.Context) error {
	_, err := sm.Insight.Acknowledge(ctx, actx, actx.EntityID, nil)
	if errors.Is(err, action.ErrStaleState) {
		return action.Fail("insight is not pending")
	}
	if err != nil {
		return fmt.Errorf("acknowledge-insight: %w", err)
	}
	return nil
}
