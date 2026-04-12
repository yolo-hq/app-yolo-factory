package actions

import (
	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

// UpdateProjectAction updates an existing project.
type UpdateProjectAction struct {
	action.SkipAllPolicies
	action.BaseUpdate[entities.Project, inputs.UpdateProjectInput]
}

func (a *UpdateProjectAction) Description() string { return "Update an existing project" }
