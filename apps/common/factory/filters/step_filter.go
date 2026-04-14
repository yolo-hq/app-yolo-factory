package filters

import (
	"github.com/yolo-hq/yolo/core/filter"
)

// StepFilter filters steps by run, phase, and status.
type StepFilter struct {
	filter.Base
	RunID  *filter.StringField `json:"run_id" filter:"run_id"`
	Phase  *filter.StringField `json:"phase" filter:"phase"`
	Status *filter.StringField `json:"status" filter:"status"`
}
