package queries

import (
	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/filter"
	"github.com/yolo-hq/yolo/core/query"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/filters"
)

func NewQuestionHandler(repo entity.ReadRepository[entities.Question]) *query.Handler[entities.Question] {
	h := query.NewHandler[entities.Question](repo)
	h.SetFilterFactory(func() filter.EntityFilter {
		return &filters.QuestionFilter{}
	})
	return h
}
