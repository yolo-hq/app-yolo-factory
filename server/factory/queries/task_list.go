package queries

import (
	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/filter"
	"github.com/yolo-hq/yolo/core/query"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/filters"
)

// NewTaskHandler creates a query handler for tasks with typed filter support.
func NewTaskHandler(repo entity.ReadRepository[entities.Task]) *query.Handler[entities.Task] {
	h := query.NewHandler[entities.Task](repo)
	h.SetFilterFactory(func() filter.EntityFilter {
		return &filters.TaskFilter{}
	})
	return h
}
