package policies

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/policy"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// CanPauseProjectData declares the entity fields this policy reads.
type CanPauseProjectData struct {
	Status string `field:"status"`
}

// CanPauseProjectPolicy denies if project status is not "active".
type CanPauseProjectPolicy struct {
	policy.EntityPolicyBase
	policy.TypedData[CanPauseProjectData]
}

func (p *CanPauseProjectPolicy) Evaluate(_ context.Context, actx *action.Context) policy.PolicyResult {
	data := p.Data(actx)
	if data.Status != entities.ProjectActive {
		return policy.Deny("project must be active")
	}
	return policy.Allow()
}
