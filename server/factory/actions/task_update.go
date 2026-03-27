package actions

import (
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/inputs"
)

type UpdateTaskAction struct {
	action.BaseUpdate[entities.Task, inputs.UpdateTaskInput]
	Repo entity.WriteRepository[entities.Task]
}
