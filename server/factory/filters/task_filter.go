package filters

import (
	"context"

	"github.com/yolo-hq/yolo/core/filter"
)

type TaskFilter struct {
	Status   *filter.StringField        `filter:"status" ops:"eq,neq,in"`
	RepoID   *filter.StringField        `filter:"repo_id" ops:"eq,in"`
	Type     *filter.StringField        `filter:"type" ops:"eq,in"`
	Priority *filter.NumberField[int]   `filter:"priority" ops:"eq,gt,gte,lt,lte"`
	ParentID *filter.StringField        `filter:"parent_id" ops:"eq"`
}

func (f *TaskFilter) Apply(ctx context.Context, fctx *filter.Context, b *filter.Builder) error {
	return filter.ApplyStruct(f, b)
}
