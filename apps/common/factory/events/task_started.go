package events

import (
	"github.com/yolo-hq/yolo/core/event"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// TaskStartedEvent is emitted when a task begins execution.
type TaskStartedEvent struct {
	event.EntityEvent[entities.Task]
}
