package policies

import (
	"context"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/policy"
)

// CanApproveSuggestionData declares the entity fields this policy reads.
type CanApproveSuggestionData struct {
	Status string `field:"status"`
}

// CanApproveSuggestionPolicy denies if suggestion status is not "pending".
type CanApproveSuggestionPolicy struct {
	policy.EntityPolicyBase
	policy.TypedData[CanApproveSuggestionData]
}

func (p *CanApproveSuggestionPolicy) Evaluate(_ context.Context, actx *action.Context) policy.PolicyResult {
	data := p.Data(actx)
	if data.Status != string(enums.SuggestionStatusPending) {
		return policy.Deny("suggestion must be pending to approve")
	}
	return policy.Allow()
}
