package actions

import (
	"github.com/yolo-hq/yolo"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/inputs"
)

type CreateQuestionAction struct {
	action.BaseCreate[entities.Question, inputs.CreateQuestionInput]
	Repo entity.WriteRepository[entities.Question]
}

func (a *CreateQuestionAction) Policies() []action.AnyPolicy {
	return []action.AnyPolicy{yolo.IsAuthenticated()}
}
