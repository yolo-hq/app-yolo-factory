package actions

import (
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/inputs"
)

type CreateQuestionAction struct {
	action.PublicAccess
	action.BaseCreate[entities.Question, inputs.CreateQuestionInput]
	Repo entity.WriteRepository[entities.Question]
}

