package policies

import (
	"context"

	"github.com/yolo-hq/yolo/core/policy"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// CanAnswerQuestionPolicy denies if question status is not "open".
type CanAnswerQuestionPolicy struct{ policy.EntityPolicyBase }

func (p *CanAnswerQuestionPolicy) PolicyData() any { return &statusData{} }
func (p *CanAnswerQuestionPolicy) EvaluateEntity(_ context.Context, data any) policy.PolicyResult {
	d := data.(*statusData)
	if d.Status != entities.QuestionOpen {
		return policy.Deny("question must be open to answer")
	}
	return policy.Allow()
}
