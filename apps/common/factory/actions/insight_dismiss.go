package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// DismissInsightAction dismisses an insight with a reason.
type DismissInsightAction struct {
	action.TypedInput[inputs.DismissInsightInput]
	action.RequirePolicy[policies.CanDismissInsightPolicy]
}

func (a *DismissInsightAction) Description() string { return "Dismiss an insight with a reason" }

func (a *DismissInsightAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	// input consumed for validation; reason not stored on entity.
	_ = a.Input(actx)

	if r := action.ExecUpdate[entities.Insight](ctx, actx, write.Set{
		write.NewField[string]("status").Value(entities.InsightDismissed),
	}); r != nil {
		return *r
	}
	actx.Resolve("Insight", actx.EntityID)
	return action.OK()
}
