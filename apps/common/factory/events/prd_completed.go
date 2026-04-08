package events

import (
	"github.com/yolo-hq/yolo/core/event"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// PRDCompletedEvent is emitted when all PRD tasks complete successfully.
type PRDCompletedEvent struct {
	event.EntityEvent[entities.PRD]
}
