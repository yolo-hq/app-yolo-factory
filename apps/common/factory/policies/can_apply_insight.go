package policies

import (
	"context"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/policy"
)

// CanApplyInsightData declares the entity fields this policy reads.
type CanApplyInsightData struct {
	Status string `field:"status"`
}

// CanApplyInsightPolicy denies if insight status is not "acknowledged".
type CanApplyInsightPolicy struct {
	policy.EntityPolicyBase
	policy.TypedData[CanApplyInsightData]
}

func (p *CanApplyInsightPolicy) Evaluate(_ context.Context, actx *action.Context) policy.PolicyResult {
	data := p.Data(actx)
	if data.Status != string(enums.InsightStatusAcknowledged) {
		return policy.Deny("insight must be acknowledged to apply")
	}
	return policy.Allow()
}
