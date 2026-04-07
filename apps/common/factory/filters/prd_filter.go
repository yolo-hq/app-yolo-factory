package filters

import (
	"context"

	"github.com/yolo-hq/yolo/core/filter"
)

// PRDFilter filters PRDs by project, status, source, and creator.
type PRDFilter struct {
	ProjectID *filter.StringField `json:"projectId" filter:"project_id"`
	Status    *filter.StringField `json:"status" filter:"status"`
	Source    *filter.StringField `json:"source" filter:"source"`
	CreatedBy *filter.StringField `json:"createdBy" filter:"created_by"`
}

func (f *PRDFilter) Apply(_ context.Context, _ *filter.Context, b *filter.Builder) error {
	return filter.ApplyStruct(f, b)
}
