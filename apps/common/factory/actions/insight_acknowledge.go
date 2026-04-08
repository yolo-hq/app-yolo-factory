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
	action.RequirePolicy[policies.CanAcknowledgeInsightPolicy]
}

func (a *AcknowledgeInsightAction) Description() string { return "Acknowledge a pending insight" }

func (a *AcknowledgeInsightAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	return action.ExecUpdate[entities.Insight](ctx, actx, write.Set{
		write.NewField[string]("status").Value(entities.InsightAcknowledged),
	})
}
