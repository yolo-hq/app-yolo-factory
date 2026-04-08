package events

import (
	"github.com/yolo-hq/yolo/core/event"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// TaskFailedEvent is emitted when a task fails.
type TaskFailedEvent struct {
	event.FailedEvent[entities.Task]
}
