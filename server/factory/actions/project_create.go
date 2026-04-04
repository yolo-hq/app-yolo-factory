package actions

import (
	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/inputs"
)

// CreateProjectAction creates a new project.
type CreateProjectAction struct {
	action.BaseCreate[entities.Project, inputs.CreateProjectInput]
}
