package policies

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/policy"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// CanAnswerQuestionData declares the entity fields this policy reads.
type CanAnswerQuestionData struct {
	Status string `field:"status"`
}

// CanAnswerQuestionPolicy denies if question status is not "open".
type CanAnswerQuestionPolicy struct {
	policy.EntityPolicyBase
	policy.TypedData[CanAnswerQuestionData]
}

func (p *CanAnswerQuestionPolicy) Evaluate(_ context.Context, actx *action.Context) policy.PolicyResult {
	data := p.Data(actx)
	if data.Status != entities.QuestionOpen {
		return policy.Deny("question must be open to answer")
	}
	return policy.Allow()
}
