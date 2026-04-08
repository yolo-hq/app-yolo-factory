package policies

import (
	"context"

	"github.com/yolo-hq/yolo/core/policy"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// CanAcknowledgeInsightPolicy denies if insight status is not "pending".
type CanAcknowledgeInsightPolicy struct{ policy.EntityPolicyBase }

func (p *CanAcknowledgeInsightPolicy) PolicyData() any { return &statusData{} }
func (p *CanAcknowledgeInsightPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.InsightPending {
		return policy.Deny("insight must be pending")
	}
	return policy.Allow()
}
