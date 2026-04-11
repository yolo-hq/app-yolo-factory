package policies

import (
	"context"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/policy"
	"github.com/yolo-hq/yolo/core/projection"
)

// CanSubmitPRDData declares the Project entity fields this policy reads.
// The runner loads the Project resolved from the input's resolves:"Project" tag.
type CanSubmitPRDData struct {
	projection.For[entities.Project]

	Status string `field:"status"`
}

// CanSubmitPRDPolicy denies if the project is not active.
type CanSubmitPRDPolicy struct {
	policy.EntityPolicyBase
	policy.Projection[CanSubmitPRDData]
}

func (p *CanSubmitPRDPolicy) Evaluate(_ context.Context, actx *action.Context) policy.PolicyResult {
	data := p.Data(actx)
	if data.Status != string(enums.ProjectStatusActive) {
		return policy.Deny("project must be active to submit a PRD")
	}
	return policy.Allow()
}
