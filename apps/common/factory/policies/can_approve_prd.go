package policies

import (
	"context"

	"github.com/yolo-hq/yolo/core/policy"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// CanApprovePRDPolicy denies if PRD status is not "draft".
type CanApprovePRDPolicy struct{ policy.EntityPolicyBase }

func (p *CanApprovePRDPolicy) PolicyData() any { return &statusData{} }
func (p *CanApprovePRDPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.PRDDraft {
		return policy.Deny("PRD must be in draft status")
	}
	return policy.Allow()
}
