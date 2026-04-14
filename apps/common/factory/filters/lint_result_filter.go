package filters

import (
	"github.com/yolo-hq/yolo/core/filter"
)

// LintResultFilter filters lint results by run, task, and passed status.
type LintResultFilter struct {
	filter.Base
	RunID  *filter.StringField `json:"run_id" filter:"run_id"`
	TaskID *filter.StringField `json:"task_id" filter:"task_id"`
	Passed *filter.StringField `json:"passed" filter:"passed"`
}
