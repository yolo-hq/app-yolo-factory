package filters

import (
	"context"

	"github.com/yolo-hq/yolo/core/filter"
)

// StepFilter filters steps by run, phase, and status.
type StepFilter struct {
	RunID  *filter.StringField `json:"runId" filter:"run_id"`
	Phase  *filter.StringField `json:"phase" filter:"phase"`
	Status *filter.StringField `json:"status" filter:"status"`
}

func (f *StepFilter) Apply(_ context.Context, _ *filter.Context, b *filter.Builder) error {
	return filter.ApplyStruct(f, b)
}
