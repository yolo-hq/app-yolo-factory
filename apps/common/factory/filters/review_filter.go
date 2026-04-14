package filters

import (
	"github.com/yolo-hq/yolo/core/filter"
)

// ReviewFilter filters reviews by run, task, and verdict.
type ReviewFilter struct {
	filter.Base
	RunID   *filter.StringField `json:"run_id" filter:"run_id"`
	TaskID  *filter.StringField `json:"task_id" filter:"task_id"`
	Verdict *filter.StringField `json:"verdict" filter:"verdict"`
}
