package actions

import (
	"context"

	"github.com/yolo-hq/yolo"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

type ResolveQuestionInput struct {
	Status     string `json:"status" validate:"required"`
	Resolution string `json:"resolution" validate:"required"`
}

type ResolveQuestionAction struct {
	action.TypedInput[ResolveQuestionInput]
	QuestionRead  entity.ReadRepository[entities.Question]
	QuestionWrite entity.WriteRepository[entities.Question]
}

func (a *ResolveQuestionAction) Policies() []action.AnyPolicy {
	return []action.AnyPolicy{yolo.IsAuthenticated()}
}

func (a *ResolveQuestionAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	input, r := a.Input(actx)
	if r != nil {
		return *r
	}

	questionID := actx.EntityID
	if questionID == "" {
		return action.Failure("question ID required")
	}

	q, _ := a.QuestionRead.FindOne(ctx, entity.FindOneOptions{ID: questionID})
	if q == nil {
		return action.NotFound("Question", questionID)
	}

	a.QuestionWrite.Update(ctx).
		Where(entity.FilterCondition{Field: "id", Operator: entity.OpEq, Value: questionID}).
		Set("status", input.Status).
		Set("resolution", input.Resolution).
		Exec(ctx)

	return action.Success(map[string]any{
		"questionId": questionID,
		"status":     input.Status,
	}, "question resolved")
}
