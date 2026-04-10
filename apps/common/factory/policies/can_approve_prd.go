package policies

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/policy"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// CanApprovePRDData declares the entity fields this policy reads.
type CanApprovePRDData struct {
	Status string `field:"status"`
}

// CanApprovePRDPolicy denies if PRD status is not "draft".
type CanApprovePRDPolicy struct {
	policy.EntityPolicyBase
	policy.TypedData[CanApprovePRDData]
}

func (p *CanApprovePRDPolicy) Evaluate(_ context.Context, actx *action.Context) policy.PolicyResult {
	data := p.Data(actx)
	if data.Status != entities.PRDDraft {
		return policy.Deny("PRD must be in draft status")
	}
	return policy.Allow()
}
