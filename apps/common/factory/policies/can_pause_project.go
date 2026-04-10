package policies

import (
	"context"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/policy"
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
	if data.Status != string(enums.ProjectStatusActive) {
		return policy.Deny("project must be active")
	}
	return policy.Allow()
}
