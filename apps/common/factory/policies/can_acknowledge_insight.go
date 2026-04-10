package policies

import (
	"context"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/policy"
)

// CanAcknowledgeInsightData declares the entity fields this policy reads.
type CanAcknowledgeInsightData struct {
	Status string `field:"status"`
}

// CanAcknowledgeInsightPolicy denies if insight status is not "pending".
type CanAcknowledgeInsightPolicy struct {
	policy.EntityPolicyBase
	policy.TypedData[CanAcknowledgeInsightData]
}

func (p *CanAcknowledgeInsightPolicy) Evaluate(_ context.Context, actx *action.Context) policy.PolicyResult {
	data := p.Data(actx)
	if data.Status != string(enums.InsightStatusPending) {
		return policy.Deny("insight must be pending")
	}
	return policy.Allow()
}
