package filters

import (
	"github.com/yolo-hq/yolo/core/filter"
)

// QuestionFilter filters questions by task, status, and confidence.
type QuestionFilter struct {
	filter.Base
	TaskID     *filter.StringField `json:"taskId" filter:"task_id"`
	Status     *filter.StringField `json:"status" filter:"status"`
	Confidence *filter.StringField `json:"confidence" filter:"confidence"`
}
