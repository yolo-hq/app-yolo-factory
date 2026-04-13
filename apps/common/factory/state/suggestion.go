package state

import (
	"github.com/yolo-hq/yolo/core/action"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
)

// Suggestion declares the lifecycle transitions for factory_suggestions.
// pending → approved → converted, or pending → rejected (terminal).
var Suggestion = action.StateMachine{
	Entity: "Suggestion",
	Field:  fields.Suggestion.Status,
	Transitions: action.Transitions{
		"Approve": {
			From: []string{string(enums.SuggestionStatusPending)},
			To:   string(enums.SuggestionStatusApproved),
		},
		"Reject": {
			From: []string{string(enums.SuggestionStatusPending)},
			To:   string(enums.SuggestionStatusRejected),
		},
		"Convert": {
			From: []string{string(enums.SuggestionStatusApproved)},
			To:   string(enums.SuggestionStatusConverted),
		},
	},
}
