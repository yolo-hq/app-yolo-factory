package policies

import (
	"context"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/policy"
	"github.com/yolo-hq/yolo/core/projection"
)

// CanArchiveProjectData declares the entity fields this policy reads.
type CanArchiveProjectData struct {
	projection.For[entities.Project]

	Status string `field:"status"`
}

// CanArchiveProjectPolicy denies if project is already archived.
type CanArchiveProjectPolicy struct {
	policy.EntityPolicyBase
	policy.Projection[CanArchiveProjectData]
}

func (p *CanArchiveProjectPolicy) Evaluate(_ context.Context, actx *action.Context) policy.PolicyResult {
	data := p.Data(actx)
	if data.Status == string(enums.ProjectStatusArchived) {
		return policy.Deny("project is already archived")
	}
	return policy.Allow()
}
