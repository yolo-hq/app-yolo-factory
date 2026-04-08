package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
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

	_, err := action.Write[entities.Insight](actx).Exec(ctx, write.Update{
		ID:  actx.EntityID,
		Set: write.Set{fields.Insight.Status.Value(entities.InsightDismissed)},
	})
	if err != nil {
		return action.Failure(err.Error())
	}
	actx.Resolve("Insight", actx.EntityID)
	return action.OK()
}
