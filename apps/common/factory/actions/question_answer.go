package actions

import (
	"context"
	"time"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// AnswerQuestionAction answers an open question.
type AnswerQuestionAction struct {
	action.TypedInput[inputs.AnswerQuestionInput]
	action.RequirePolicy[policies.CanAnswerQuestionPolicy]
}

func (a *AnswerQuestionAction) Description() string { return "Answer an open question" }

func (a *AnswerQuestionAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	input := a.Input(actx)
	now := time.Now()

	_, err := action.Write[entities.Question](actx).Exec(ctx, write.Update{
		ID: actx.EntityID,
		Set: write.Set{
			fields.Question.Status.Value(string(enums.QuestionStatusAnswered)),
			fields.Question.Answer.Value(input.Answer),
			fields.Question.AnsweredBy.Value("human"),
			fields.Question.AnsweredAt.Value(&now),
		},
	})
	if err != nil {
		return action.Failure(err.Error())
	}

	actx.Resolve("Question", actx.EntityID)
	return action.OK()
}
