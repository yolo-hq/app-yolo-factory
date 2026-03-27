package actions

import (
	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/inputs"
)

type ResolveQuestionAction struct {
	action.BaseUpdate[entities.Question, inputs.ResolveQuestionInput]
}
