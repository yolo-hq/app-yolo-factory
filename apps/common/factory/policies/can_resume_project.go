package policies

import (
	"context"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/policy"
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
	if data.Status != string(enums.ProjectStatusPaused) {
		return policy.Deny("project must be paused to resume")
	}
	return policy.Allow()
}
