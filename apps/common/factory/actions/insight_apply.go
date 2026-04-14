package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// ApplyInsightAction applies an acknowledged insight.
type ApplyInsightAction struct {
	action.RequirePolicy[policies.CanApplyInsightPolicy]
	action.NoInput
}

func (a *ApplyInsightAction) Description() string { return "Apply an acknowledged insight" }

func (a *ApplyInsightAction) Execute(ctx context.Context, actx *action.Context) error {
	_, err := sm.Insight.Apply(ctx, actx, actx.EntityID, nil)
	if errors.Is(err, action.ErrStaleState) {
		return action.Fail("insight is not acknowledged")
	}
	if err != nil {
		return fmt.Errorf("apply-insight: %w", err)
	}
	return nil
}
