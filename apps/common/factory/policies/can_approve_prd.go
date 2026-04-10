package policies

import (
	"context"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/policy"
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
	if data.Status != string(enums.PRDStatusDraft) {
		return policy.Deny("PRD must be in draft status")
	}
	return policy.Allow()
}
