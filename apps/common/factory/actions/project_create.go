package actions

import (
	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

// CreateProjectAction creates a new project.
type CreateProjectAction struct {
	action.BaseCreate[entities.Project, inputs.CreateProjectInput]
	action.SkipAllPolicies
}

func (a *CreateProjectAction) Description() string { return "Create a new project" }
