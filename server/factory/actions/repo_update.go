package actions

import (
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/inputs"
)

type UpdateRepoAction struct {
	action.BaseUpdate[entities.Repo, inputs.UpdateRepoInput]
	Repo entity.WriteRepository[entities.Repo]
}
