package state

import (
	"github.com/yolo-hq/yolo/core/action"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
)

// Project declares the lifecycle transitions for factory_projects.
// active <-> paused; either can be archived (terminal).
var Project = action.StateMachine{
	Entity: "Project",
	Field:  fields.Project.Status,
	Transitions: action.Transitions{
		"Pause": {
			From: []string{string(enums.ProjectStatusActive)},
			To:   string(enums.ProjectStatusPaused),
		},
		"Resume": {
			From: []string{string(enums.ProjectStatusPaused)},
			To:   string(enums.ProjectStatusActive),
		},
		"Archive": {
			From: []string{string(enums.ProjectStatusActive), string(enums.ProjectStatusPaused)},
			To:   string(enums.ProjectStatusArchived),
		},
	},
}
