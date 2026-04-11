package policies

import (
	"context"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/policy"
	"github.com/yolo-hq/yolo/core/projection"
)

// CanAnswerQuestionData declares the entity fields this policy reads.
type CanAnswerQuestionData struct {
	projection.For[entities.Question]

	Status string `field:"status"`
}

// CanAnswerQuestionPolicy denies if question status is not "open".
type CanAnswerQuestionPolicy struct {
	policy.EntityPolicyBase
	policy.Projection[CanAnswerQuestionData]
}

func (p *CanAnswerQuestionPolicy) Evaluate(_ context.Context, actx *action.Context) policy.PolicyResult {
	data := p.Data(actx)
	if data.Status != string(enums.QuestionStatusOpen) {
		return policy.Deny("question must be open to answer")
	}
	return policy.Allow()
}
