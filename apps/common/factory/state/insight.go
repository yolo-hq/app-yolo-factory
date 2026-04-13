package state

import (
	"github.com/yolo-hq/yolo/core/action"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
)

// Insight declares the lifecycle transitions for factory_insights.
// pending → acknowledged → applied; pending|acknowledged → dismissed (terminal).
var Insight = action.StateMachine{
	Entity: "Insight",
	Field:  fields.Insight.Status,
	Transitions: action.Transitions{
		"Acknowledge": {
			From: []string{string(enums.InsightStatusPending)},
			To:   string(enums.InsightStatusAcknowledged),
		},
		"Apply": {
			From: []string{string(enums.InsightStatusAcknowledged)},
			To:   string(enums.InsightStatusApplied),
		},
		"Dismiss": {
			From: []string{string(enums.InsightStatusPending), string(enums.InsightStatusAcknowledged)},
			To:   string(enums.InsightStatusDismissed),
		},
	},
}
