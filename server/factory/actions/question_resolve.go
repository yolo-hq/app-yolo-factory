package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/inputs"
)

type ResolveQuestionAction struct {
	action.TypedInput[inputs.ResolveQuestionInput]
	QuestionRead  entity.ReadRepository[entities.Question]
	QuestionWrite entity.WriteRepository[entities.Question]
}


func (a *ResolveQuestionAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	input := a.Get(actx)

	// EntityID comes from URL path — validated by framework
	questionID := actx.EntityID

	a.QuestionWrite.Update(ctx).
		WhereID(questionID).
		Set("status", input.Status).
		Set("resolution", input.Resolution).
		Exec(ctx)

	actx.Resolve("Question", questionID)
	return action.OK()
}
