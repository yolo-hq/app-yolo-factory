package policies

import (
	"context"
	"fmt"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/policy"
)

// CanExecutePRDData declares the entity fields this policy reads.
type CanExecutePRDData struct {
	Status string `field:"status"`
}

// CanExecutePRDPolicy denies if PRD status is not "draft" or "approved".
type CanExecutePRDPolicy struct {
	policy.EntityPolicyBase
	policy.TypedData[CanExecutePRDData]
}

func (p *CanExecutePRDPolicy) Evaluate(_ context.Context, actx *action.Context) policy.PolicyResult {
	data := p.Data(actx)
	if data.Status != string(enums.PRDStatusDraft) && data.Status != string(enums.PRDStatusApproved) {
		return policy.Deny(fmt.Sprintf("PRD must be in draft or approved status, got %q", data.Status))
	}
	return policy.Allow()
}
