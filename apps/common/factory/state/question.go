package state

import (
	"github.com/yolo-hq/yolo/core/action"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
)

// Question declares the lifecycle transitions for factory_questions.
// open → answered (human) | auto_resolved (system).
var Question = action.StateMachine{
	Entity: "Question",
	Field:  fields.Question.Status,
	Transitions: action.Transitions{
		"Answer": {
			From: []string{string(enums.QuestionStatusOpen)},
			To:   string(enums.QuestionStatusAnswered),
		},
		"AutoResolve": {
			From: []string{string(enums.QuestionStatusOpen)},
			To:   string(enums.QuestionStatusAutoResolved),
		},
	},
}
