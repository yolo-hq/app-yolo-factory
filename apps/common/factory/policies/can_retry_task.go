package policies

import (
	"context"

	"github.com/yolo-hq/yolo/core/policy"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// CanRetryTaskPolicy denies if task status is not "failed".
type CanRetryTaskPolicy struct{ policy.EntityPolicyBase }

func (p *CanRetryTaskPolicy) PolicyData() any { return &statusData{} }
func (p *CanRetryTaskPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.TaskFailed {
		return policy.Deny("task must be failed to retry")
	}
	return policy.Allow()
}
