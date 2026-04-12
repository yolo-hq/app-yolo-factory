package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/.yolo/repos"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// DismissInsightAction dismisses an insight.
type DismissInsightAction struct {
	action.RequirePolicy[policies.CanDismissInsightPolicy]
}

func (a *DismissInsightAction) Description() string { return "Dismiss an insight" }

func (a *DismissInsightAction) Execute(ctx context.Context, actx *action.Context) error {
	_, err := repos.Insight.UpdateEntity(ctx, actx, write.Set{
		fields.Insight.Status.Value(string(enums.InsightStatusDismissed)),
	})
	return err
}
