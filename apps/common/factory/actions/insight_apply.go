package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// ApplyInsightAction applies an acknowledged insight.
type ApplyInsightAction struct {
	action.NoInput
	action.RequirePolicy[policies.CanApplyInsightPolicy]
}

func (a *ApplyInsightAction) Description() string { return "Apply an acknowledged insight" }

func (a *ApplyInsightAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	if r := action.ExecUpdate[entities.Insight](ctx, actx, write.Set{
		write.NewField[string]("status").Value(entities.InsightApplied),
	}); r != nil {
		return *r
	}
	actx.Resolve("Insight", actx.EntityID)
	return action.OK()
}
