package actions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/constants"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

// QuestionAnswer answers an open question.
//
// @policy CanAnswerQuestionPolicy
func QuestionAnswer(ctx context.Context, actx *action.Context, in inputs.AnswerQuestionInput) error {
	now := time.Now()
	_, err := sm.Question.Answer(ctx, actx, actx.EntityID, write.Set{
		fields.Question.Answer.Value(in.Answer),
		fields.Question.AnsweredBy.Value(constants.ActorHuman),
		fields.Question.AnsweredAt.Value(&now),
	})
	if errors.Is(err, action.ErrStaleState) {
		return action.Fail("question is not open")
	}
	if err != nil {
		return fmt.Errorf("answer-question: %w", err)
	}
	return nil
}
