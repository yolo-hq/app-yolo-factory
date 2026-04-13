package state

import (
	"github.com/yolo-hq/yolo/core/action"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
)

// Run declares the lifecycle transitions for factory_runs.
// running → completed|failed|cancelled|blocked. Blocked can resume to running.
var Run = action.StateMachine{
	Entity: "Run",
	Field:  fields.Run.Status,
	Transitions: action.Transitions{
		"Complete": {
			From: []string{string(enums.RunStatusRunning)},
			To:   string(enums.RunStatusCompleted),
		},
		"Fail": {
			From: []string{string(enums.RunStatusRunning), string(enums.RunStatusBlocked)},
			To:   string(enums.RunStatusFailed),
		},
		"Cancel": {
			From: []string{string(enums.RunStatusRunning), string(enums.RunStatusBlocked)},
			To:   string(enums.RunStatusCancelled),
		},
		"Block": {
			From: []string{string(enums.RunStatusRunning)},
			To:   string(enums.RunStatusBlocked),
		},
		"Resume": {
			From: []string{string(enums.RunStatusBlocked)},
			To:   string(enums.RunStatusRunning),
		},
	},
}
