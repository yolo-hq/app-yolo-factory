package actions

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// DismissInsightAction dismisses an insight.
type DismissInsightAction struct {
	action.RequirePolicy[policies.CanDismissInsightPolicy]
}

func (a *DismissInsightAction) Description() string { return "Dismiss an insight" }

func (a *DismissInsightAction) Execute(ctx context.Context, actx *action.Context) error {
	_, err := sm.Insight.Dismiss(ctx, actx, actx.EntityID, nil)
	if err != nil {
		return fmt.Errorf("dismiss-insight: %w", err)
	}
	return nil
}
