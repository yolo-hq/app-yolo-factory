package policies

import (
	"context"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/policy"
)

// CanSubmitPRDData declares the Project entity fields this policy reads.
// The runner loads the Project resolved from the input's resolves:"Project" tag.
type CanSubmitPRDData struct {
	Status string `field:"status"`
}

// CanSubmitPRDPolicy denies if the project is not active.
type CanSubmitPRDPolicy struct {
	policy.EntityPolicyBase
	policy.TypedData[CanSubmitPRDData]
}

func (p *CanSubmitPRDPolicy) Evaluate(_ context.Context, actx *action.Context) policy.PolicyResult {
	data := p.Data(actx)
	if data.Status != string(enums.ProjectStatusActive) {
		return policy.Deny("project must be active to submit a PRD")
	}
	return policy.Allow()
}
