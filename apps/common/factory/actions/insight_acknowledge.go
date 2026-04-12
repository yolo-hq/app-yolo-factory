package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// AcknowledgeInsightAction acknowledges a pending insight.
type AcknowledgeInsightAction struct {
	action.RequirePolicy[policies.CanAcknowledgeInsightPolicy]
	action.NoInput
}

func (a *AcknowledgeInsightAction) Description() string { return "Acknowledge a pending insight" }

func (a *AcknowledgeInsightAction) Execute(ctx context.Context, actx *action.Context) error {
	_, err := action.Write[entities.Insight](actx).Exec(ctx, write.Update{
		ID:  actx.EntityID,
		Set: write.Set{fields.Insight.Status.Value(string(enums.InsightStatusAcknowledged))},
	})
	return err
}
