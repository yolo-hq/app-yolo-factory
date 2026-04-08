package policies

import (
	"context"

	"github.com/yolo-hq/yolo/core/policy"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// CanApplyInsightPolicy denies if insight status is not "acknowledged".
type CanApplyInsightPolicy struct{ policy.EntityPolicyBase }

func (p *CanApplyInsightPolicy) PolicyData() any { return &statusData{} }
func (p *CanApplyInsightPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.InsightAcknowledged {
		return policy.Deny("insight must be acknowledged to apply")
	}
	return policy.Allow()
}
