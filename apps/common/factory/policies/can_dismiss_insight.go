package policies

import (
	"context"
	"fmt"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/policy"
	"github.com/yolo-hq/yolo/core/projection"
)

// CanDismissInsightData declares the entity fields this policy reads.
type CanDismissInsightData struct {
	projection.For[entities.Insight]

	Status string `field:"status"`
}

// CanDismissInsightPolicy denies if insight is not "pending" or "acknowledged".
type CanDismissInsightPolicy struct {
	policy.EntityPolicyBase
	policy.Projection[CanDismissInsightData]
}

func (p *CanDismissInsightPolicy) Evaluate(_ context.Context, actx *action.Context) policy.PolicyResult {
	data := p.Data(actx)
	if data.Status != string(enums.InsightStatusPending) && data.Status != string(enums.InsightStatusAcknowledged) {
		return policy.Deny(fmt.Sprintf("insight must be pending or acknowledged to dismiss, got %q", data.Status))
	}
	return policy.Allow()
}
