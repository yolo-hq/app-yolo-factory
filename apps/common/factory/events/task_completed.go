package events

import (
	"github.com/yolo-hq/yolo/core/event"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// TaskCompletedEvent is emitted when a task completes successfully.
type TaskCompletedEvent struct {
	event.EntityEvent[entities.Task]
}
