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

// ApplyInsightAction applies an acknowledged insight.
type ApplyInsightAction struct {
	action.RequirePolicy[policies.CanApplyInsightPolicy]
	action.NoInput
}

func (a *ApplyInsightAction) Description() string { return "Apply an acknowledged insight" }

func (a *ApplyInsightAction) Execute(ctx context.Context, actx *action.Context) error {
	_, err := repos.Insight.UpdateEntity(ctx, actx, write.Set{
		fields.Insight.Status.Value(string(enums.InsightStatusApplied)),
	})
	return err
}
