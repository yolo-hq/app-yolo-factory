package events

import (
	"github.com/yolo-hq/yolo/core/event"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// PRDPlanningCompleteEvent is emitted when PRD planning finishes.
type PRDPlanningCompleteEvent struct {
	event.EntityEvent[entities.PRD]
}
