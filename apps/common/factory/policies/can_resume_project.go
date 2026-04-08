package policies

import (
	"context"

	"github.com/yolo-hq/yolo/core/policy"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// CanResumeProjectPolicy denies if project status is not "paused".
type CanResumeProjectPolicy struct{ policy.EntityPolicyBase }

func (p *CanResumeProjectPolicy) PolicyData() any { return &statusData{} }
func (p *CanResumeProjectPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.ProjectPaused {
		return policy.Deny("project must be paused to resume")
	}
	return policy.Allow()
}
