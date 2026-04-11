package policies

import (
	"context"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/policy"
	"github.com/yolo-hq/yolo/core/projection"
)

// CanPauseProjectData declares the entity fields this policy reads.
type CanPauseProjectData struct {
	projection.For[entities.Project]

	Status string `field:"status"`
}

// CanPauseProjectPolicy denies if project status is not "active".
type CanPauseProjectPolicy struct {
	policy.EntityPolicyBase
	policy.Projection[CanPauseProjectData]
}

func (p *CanPauseProjectPolicy) Evaluate(_ context.Context, actx *action.Context) policy.PolicyResult {
	data := p.Data(actx)
	if data.Status != string(enums.ProjectStatusActive) {
		return policy.Deny("project must be active")
	}
	return policy.Allow()
}
