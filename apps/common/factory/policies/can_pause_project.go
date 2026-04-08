package policies

import (
	"context"

	"github.com/yolo-hq/yolo/core/policy"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// CanPauseProjectPolicy denies if project status is not "active".
type CanPauseProjectPolicy struct{ policy.EntityPolicyBase }

func (p *CanPauseProjectPolicy) PolicyData() any { return &statusData{} }
func (p *CanPauseProjectPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.ProjectActive {
		return policy.Deny("project must be active")
	}
	return policy.Allow()
}
