package policies

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/policy"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// CanExecutePRDPolicy denies if PRD status is not "draft" or "approved".
type CanExecutePRDPolicy struct{ policy.EntityPolicyBase }

func (p *CanExecutePRDPolicy) PolicyData() any { return &statusData{} }
func (p *CanExecutePRDPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.PRDDraft && d.Status != entities.PRDApproved {
		return policy.Deny(fmt.Sprintf("PRD must be in draft or approved status, got %q", d.Status))
	}
	return policy.Allow()
}
