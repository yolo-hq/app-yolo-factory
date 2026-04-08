package policies

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/policy"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// CanCancelTaskPolicy denies if task is in a terminal state (done, failed, cancelled).
type CanCancelTaskPolicy struct{ policy.EntityPolicyBase }

func (p *CanCancelTaskPolicy) PolicyData() any { return &statusData{} }
func (p *CanCancelTaskPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	switch d.Status {
	case entities.TaskDone, entities.TaskFailed, entities.TaskCancelled:
		return policy.Deny(fmt.Sprintf("task cannot be cancelled in %q status", d.Status))
	}
	return policy.Allow()
}
