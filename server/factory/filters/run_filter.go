package filters

import (
	"context"

	"github.com/yolo-hq/yolo/core/filter"
)

type RunFilter struct {
	TaskID *filter.StringField `filter:"task_id" ops:"eq,in"`
	RepoID *filter.StringField `filter:"repo_id" ops:"eq,in"`
	Status *filter.StringField `filter:"status" ops:"eq,neq,in"`
	Agent  *filter.StringField `filter:"agent" ops:"eq,in"`
}

func (f *RunFilter) Apply(ctx context.Context, fctx *filter.Context, b *filter.Builder) error {
	return filter.ApplyStruct(f, b)
}
