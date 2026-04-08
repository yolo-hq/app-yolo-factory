package policies

import (
	"context"

	"github.com/yolo-hq/yolo/core/policy"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// CanApproveSuggestionPolicy denies if suggestion status is not "pending".
type CanApproveSuggestionPolicy struct{ policy.EntityPolicyBase }

func (p *CanApproveSuggestionPolicy) PolicyData() any { return &statusData{} }
func (p *CanApproveSuggestionPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.SuggestionPending {
		return policy.Deny("suggestion must be pending to approve")
	}
	return policy.Allow()
}
