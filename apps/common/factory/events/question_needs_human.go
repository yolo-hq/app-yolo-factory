package events

import (
	"github.com/yolo-hq/yolo/core/event"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// QuestionNeedsHumanEvent is emitted when a question requires human input.
type QuestionNeedsHumanEvent struct {
	event.EntityEvent[entities.Question]
}
