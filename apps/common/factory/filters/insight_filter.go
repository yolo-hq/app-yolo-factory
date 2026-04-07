package filters

import (
	"context"

	"github.com/yolo-hq/yolo/core/filter"
)

// InsightFilter filters insights by project, category, status, and priority.
type InsightFilter struct {
	ProjectID *filter.StringField `json:"projectId" filter:"project_id"`
	Category  *filter.StringField `json:"category" filter:"category"`
	Status    *filter.StringField `json:"status" filter:"status"`
	Priority  *filter.StringField `json:"priority" filter:"priority"`
}

func (f *InsightFilter) Apply(_ context.Context, _ *filter.Context, b *filter.Builder) error {
	return filter.ApplyStruct(f, b)
}
