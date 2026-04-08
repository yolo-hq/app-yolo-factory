package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// AcknowledgeInsightAction acknowledges a pending insight.
type AcknowledgeInsightAction struct {
	action.NoInput
}

func (a *AcknowledgeInsightAction) Policies() []action.AnyPolicy {
	return []action.AnyPolicy{&policies.InsightMustBePending{}}
}

func (a *AcknowledgeInsightAction) Execute(ctx context.Context, actx *action.Context) action.Result {
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
