package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/inputs"
)

// DismissInsightAction dismisses an insight with a reason.
type DismissInsightAction struct {
	action.TypedInput[inputs.DismissInsightInput]
}

func (a *DismissInsightAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	_, r := action.FindOrFail[entities.Insight](ctx, action.ReadRepo[entities.Insight](actx), actx.EntityID)
	if r != nil {
		return *r
	}

	// input consumed for validation; reason not stored on entity.
	_ = a.Input(actx)

	_, err := action.Write[entities.Insight](actx).Exec(ctx, write.Update{
		ID:  actx.EntityID,
		Set: write.Set{write.NewField[string]("status").Value(entities.InsightDismissed)},
	})
	if err != nil {
		return action.Failure(err.Error())
	}

	actx.Resolve("Insight", actx.EntityID)
	return action.OK()
}
