package policies

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/policy"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// CanRetryTaskData declares the entity fields this policy reads.
type CanRetryTaskData struct {
	Status string `field:"status"`
}

// CanRetryTaskPolicy denies if task status is not "failed".
type CanRetryTaskPolicy struct {
	policy.EntityPolicyBase
	policy.TypedData[CanRetryTaskData]
}

func (p *CanRetryTaskPolicy) Evaluate(_ context.Context, actx *action.Context) policy.PolicyResult {
	data := p.Data(actx)
	if data.Status != entities.TaskFailed {
		return policy.Deny("task must be failed to retry")
	}
	return policy.Allow()
}
