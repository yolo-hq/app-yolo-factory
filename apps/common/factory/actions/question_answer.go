package actions

import (
	"context"
	"time"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// AnswerQuestionAction answers an open question.
type AnswerQuestionAction struct {
	action.TypedInput[inputs.AnswerQuestionInput]
}

func (a *AnswerQuestionAction) Policies() []action.AnyPolicy {
	return []action.AnyPolicy{&policies.QuestionMustBeOpen{}}
}

func (a *AnswerQuestionAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	input := a.Input(actx)
	now := time.Now()

	_, err := action.Write[entities.Question](actx).Exec(ctx, write.Update{
		ID: actx.EntityID,
		Set: write.Set{
			write.NewField[string]("status").Value(entities.QuestionAnswered),
			write.NewField[string]("answer").Value(input.Answer),
			write.NewField[string]("answered_by").Value("human"),
			write.NewField[*time.Time]("answered_at").Value(&now),
		},
	})
	if err != nil {
		return action.Failure(err.Error())
	}

	actx.Resolve("Question", actx.EntityID)
	return action.OK()
}
