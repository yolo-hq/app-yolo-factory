package actions

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// AcknowledgeInsightAction acknowledges a pending insight.
type AcknowledgeInsightAction struct {
	action.RequirePolicy[policies.CanAcknowledgeInsightPolicy]
	action.NoInput
}

func (a *AcknowledgeInsightAction) Description() string { return "Acknowledge a pending insight" }

func (a *AcknowledgeInsightAction) Execute(ctx context.Context, actx *action.Context) error {
	_, err := sm.Insight.Acknowledge(ctx, actx, actx.EntityID, nil)
	if err != nil {
		return fmt.Errorf("acknowledge-insight: %w", err)
	}
	return nil
}
