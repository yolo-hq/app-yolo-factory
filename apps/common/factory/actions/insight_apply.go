package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// ApplyInsightAction applies an acknowledged insight.
type ApplyInsightAction struct {
	action.NoInput
}

func (a *ApplyInsightAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	insight, r := action.FindOrFail[entities.Insight](ctx, action.ReadRepo[entities.Insight](actx), actx.EntityID)
	if r != nil {
		return *r
	}

	if insight.Status != entities.InsightAcknowledged {
		return action.Failure("insight must be acknowledged to apply")
	}

	_, err := action.Write[entities.Insight](actx).Exec(ctx, write.Update{
		ID:  actx.EntityID,
		Set: write.Set{write.NewField[string]("status").Value(entities.InsightApplied)},
	})
	if err != nil {
		return action.Failure(err.Error())
	}

	actx.Resolve("Insight", actx.EntityID)
	return action.OK()
}
