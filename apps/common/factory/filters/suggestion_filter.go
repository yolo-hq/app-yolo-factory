package filters

import (
	"context"

	"github.com/yolo-hq/yolo/core/filter"
)

// SuggestionFilter filters suggestions by project, source, category, status, and priority.
type SuggestionFilter struct {
	ProjectID *filter.StringField `json:"projectId" filter:"project_id"`
	Source    *filter.StringField `json:"source" filter:"source"`
	Category  *filter.StringField `json:"category" filter:"category"`
	Status    *filter.StringField `json:"status" filter:"status"`
	Priority  *filter.StringField `json:"priority" filter:"priority"`
}

func (f *SuggestionFilter) Apply(_ context.Context, _ *filter.Context, b *filter.Builder) error {
	return filter.ApplyStruct(f, b)
}
