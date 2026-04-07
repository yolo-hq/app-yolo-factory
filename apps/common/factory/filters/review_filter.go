package filters

import (
	"context"

	"github.com/yolo-hq/yolo/core/filter"
)

// ReviewFilter filters reviews by run, task, and verdict.
type ReviewFilter struct {
	RunID   *filter.StringField `json:"runId" filter:"run_id"`
	TaskID  *filter.StringField `json:"taskId" filter:"task_id"`
	Verdict *filter.StringField `json:"verdict" filter:"verdict"`
}

func (f *ReviewFilter) Apply(_ context.Context, _ *filter.Context, b *filter.Builder) error {
	return filter.ApplyStruct(f, b)
}
