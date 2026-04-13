package state

import (
	"github.com/yolo-hq/yolo/core/action"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/events"
)

// PRD declares the lifecycle transitions for factory_prds.
// draft → approved → planning → in_progress → completed|failed.
// Approve and Execute can both originate from draft for legacy flows.
var PRD = action.StateMachine{
	Entity: "PRD",
	Field:  fields.PRD.Status,
	Transitions: action.Transitions{
		"Approve": {
			From: []string{string(enums.PRDStatusDraft)},
			To:   string(enums.PRDStatusApproved),
		},
		"Execute": {
			From: []string{string(enums.PRDStatusDraft), string(enums.PRDStatusApproved)},
			To:   string(enums.PRDStatusPlanning),
		},
		"StartProgress": {
			From: []string{string(enums.PRDStatusPlanning)},
			To:   string(enums.PRDStatusInProgress),
		},
		"Complete": {
			From:  []string{string(enums.PRDStatusInProgress), string(enums.PRDStatusPlanning)},
			To:    string(enums.PRDStatusCompleted),
			Event: events.PRDCompletedName,
		},
		"Fail": {
			From:  []string{string(enums.PRDStatusInProgress), string(enums.PRDStatusPlanning), string(enums.PRDStatusDraft), string(enums.PRDStatusApproved)},
			To:    string(enums.PRDStatusFailed),
			Event: events.PRDFailedName,
		},
	},
}
