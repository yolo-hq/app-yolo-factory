package policies

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/policy"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
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
	if data.Status != entities.PRDDraft && data.Status != entities.PRDApproved {
		return policy.Deny(fmt.Sprintf("PRD must be in draft or approved status, got %q", data.Status))
	}
	return policy.Allow()
}
