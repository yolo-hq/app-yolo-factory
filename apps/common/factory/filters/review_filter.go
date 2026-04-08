package filters

import (
	"github.com/yolo-hq/yolo/core/filter"
)

// ReviewFilter filters reviews by run, task, and verdict.
type ReviewFilter struct {
	filter.Base
	RunID   *filter.StringField `json:"runId" filter:"run_id"`
	TaskID  *filter.StringField `json:"taskId" filter:"task_id"`
	Verdict *filter.StringField `json:"verdict" filter:"verdict"`
}
