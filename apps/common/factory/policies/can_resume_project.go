package policies

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/policy"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// CanResumeProjectData declares the entity fields this policy reads.
type CanResumeProjectData struct {
	Status string `field:"status"`
}

// CanResumeProjectPolicy denies if project status is not "paused".
type CanResumeProjectPolicy struct {
	policy.EntityPolicyBase
	policy.TypedData[CanResumeProjectData]
}

func (p *CanResumeProjectPolicy) Evaluate(_ context.Context, actx *action.Context) policy.PolicyResult {
	data := p.Data(actx)
	if data.Status != entities.ProjectPaused {
		return policy.Deny("project must be paused to resume")
	}
	return policy.Allow()
}
