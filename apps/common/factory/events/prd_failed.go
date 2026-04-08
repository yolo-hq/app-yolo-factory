package events

import (
	"github.com/yolo-hq/yolo/core/event"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// PRDFailedEvent is emitted when a PRD fails.
type PRDFailedEvent struct {
	event.FailedEvent[entities.PRD]
}
