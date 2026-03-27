package actions

import (
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/inputs"
)

type CreateRunAction struct {
	action.PublicAccess
	action.BaseCreate[entities.Run, inputs.CreateRunInput]
	Repo entity.WriteRepository[entities.Run]
}

