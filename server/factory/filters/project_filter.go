package filters

import (
	"context"

	"github.com/yolo-hq/yolo/core/filter"
)

// ProjectFilter filters projects by status and name.
type ProjectFilter struct {
	Status *filter.StringField `json:"status" filter:"status"`
	Name   *filter.StringField `json:"name" filter:"name"`
}

func (f *ProjectFilter) Apply(_ context.Context, _ *filter.Context, b *filter.Builder) error {
	return filter.ApplyStruct(f, b)
}
