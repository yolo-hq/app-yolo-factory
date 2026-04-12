package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/.yolo/repos"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/constants"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// AnswerQuestionAction answers an open question.
type AnswerQuestionAction struct {
	action.RequirePolicy[policies.CanAnswerQuestionPolicy]
	action.TypedInput[inputs.AnswerQuestionInput]
}

func (a *AnswerQuestionAction) Description() string { return "Answer an open question" }

func (a *AnswerQuestionAction) Execute(ctx context.Context, actx *action.Context) error {
	input := a.Input(actx)
	now := time.Now()

	_, err := repos.Question.UpdateEntity(ctx, actx, write.Set{
		fields.Question.Status.Value(string(enums.QuestionStatusAnswered)),
		fields.Question.Answer.Value(input.Answer),
		fields.Question.AnsweredBy.Value(constants.ActorHuman),
		fields.Question.AnsweredAt.Value(&now),
	})
	if err != nil {
		return fmt.Errorf("answer-question: %w", err)
	}
	return nil
}
