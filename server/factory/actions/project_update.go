package actions

import (
	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/inputs"
)

// UpdateProjectAction updates an existing project.
type UpdateProjectAction struct {
	action.BaseUpdate[entities.Project, inputs.UpdateProjectInput]
}
