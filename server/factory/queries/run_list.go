package queries

import (
	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/filter"
	"github.com/yolo-hq/yolo/core/query"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/filters"
)

// NewRunHandler creates a query handler for runs with typed filter support.
func NewRunHandler(repo entity.ReadRepository[entities.Run]) *query.Handler[entities.Run] {
	h := query.NewHandler[entities.Run](repo)
	h.SetFilterFactory(func() filter.EntityFilter {
		return &filters.RunFilter{}
	})
	return h
}
