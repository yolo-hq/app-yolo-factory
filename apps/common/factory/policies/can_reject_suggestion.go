package policies

import (
	"context"

	"github.com/yolo-hq/yolo/core/policy"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// CanRejectSuggestionPolicy denies if suggestion status is not "pending".
type CanRejectSuggestionPolicy struct{ policy.EntityPolicyBase }

func (p *CanRejectSuggestionPolicy) PolicyData() any { return &statusData{} }
func (p *CanRejectSuggestionPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.SuggestionPending {
		return policy.Deny("suggestion must be pending to reject")
	}
	return policy.Allow()
}
