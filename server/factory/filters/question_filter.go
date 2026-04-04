package filters

import (
	"context"

	"github.com/yolo-hq/yolo/core/filter"
)

// QuestionFilter filters questions by task, status, and confidence.
type QuestionFilter struct {
	TaskID     *filter.StringField `json:"taskId" filter:"task_id"`
	Status     *filter.StringField `json:"status" filter:"status"`
	Confidence *filter.StringField `json:"confidence" filter:"confidence"`
}

func (f *QuestionFilter) Apply(_ context.Context, _ *filter.Context, b *filter.Builder) error {
	return filter.ApplyStruct(f, b)
}
