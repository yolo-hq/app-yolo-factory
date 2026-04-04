package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

// AcknowledgeInsightAction acknowledges a pending insight.
type AcknowledgeInsightAction struct {
	action.NoInput
}

func (a *AcknowledgeInsightAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	insight, r := action.FindOrFail[entities.Insight](ctx, action.ReadRepo[entities.Insight](actx), actx.EntityID)
	if r != nil {
		return *r
	}

	if insight.Status != entities.InsightPending {
		return action.Failure("insight must be pending to acknowledge")
	}

	_, err := action.Write[entities.Insight](actx).Exec(ctx, write.Update{
		ID:  actx.EntityID,
		Set: write.Set{write.NewField[string]("status").Value(entities.InsightAcknowledged)},
	})
	if err != nil {
		return action.Failure(err.Error())
	}

	actx.Resolve("Insight", actx.EntityID)
	return action.OK()
}
