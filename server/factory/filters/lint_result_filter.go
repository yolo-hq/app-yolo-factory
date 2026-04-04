package filters

import (
	"context"

	"github.com/yolo-hq/yolo/core/filter"
)

// LintResultFilter filters lint results by run, task, and passed status.
type LintResultFilter struct {
	RunID  *filter.StringField `json:"runId" filter:"run_id"`
	TaskID *filter.StringField `json:"taskId" filter:"task_id"`
	Passed *filter.StringField `json:"passed" filter:"passed"`
}

func (f *LintResultFilter) Apply(_ context.Context, _ *filter.Context, b *filter.Builder) error {
	return filter.ApplyStruct(f, b)
}
