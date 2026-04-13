// Package state declares action.StateMachine vars for factory entities.
// Each var is auto-discovered by yolo codegen and exposed via .yolo/sm/sm.go
// as a typed wrapper with one method per transition.
package state

import (
	"github.com/yolo-hq/yolo/core/action"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/events"
)

// Task declares the lifecycle transitions for factory_tasks.
// Cancel allows queued/blocked/running/reviewing — every non-terminal state.
// Complete/Fail come from running (or reviewing). Retry resets failed → queued.
var Task = action.StateMachine{
	Entity: "Task",
	Field:  fields.Task.Status,
	Transitions: action.Transitions{
		"Complete": {
			From:  []string{string(enums.TaskStatusRunning), string(enums.TaskStatusReviewing)},
			To:    string(enums.TaskStatusDone),
			Event: events.TaskCompletedName,
		},
		"Fail": {
			From:  []string{string(enums.TaskStatusRunning), string(enums.TaskStatusReviewing)},
			To:    string(enums.TaskStatusFailed),
			Event: events.TaskFailedName,
		},
		"Retry": {
			From: []string{string(enums.TaskStatusFailed)},
			To:   string(enums.TaskStatusQueued),
		},
		"Requeue": {
			From: []string{string(enums.TaskStatusRunning), string(enums.TaskStatusReviewing)},
			To:   string(enums.TaskStatusQueued),
		},
		"Cancel": {
			From: []string{
				string(enums.TaskStatusQueued),
				string(enums.TaskStatusBlocked),
				string(enums.TaskStatusRunning),
				string(enums.TaskStatusReviewing),
			},
			To: string(enums.TaskStatusCancelled),
		},
		"Block": {
			From:  []string{string(enums.TaskStatusQueued), string(enums.TaskStatusRunning)},
			To:    string(enums.TaskStatusBlocked),
			Event: events.TaskBlockedName,
		},
		"Unblock": {
			From: []string{string(enums.TaskStatusBlocked)},
			To:   string(enums.TaskStatusQueued),
		},
	},
}
