package filters

import (
	"github.com/yolo-hq/yolo/core/filter"
)

// InsightFilter filters insights by project, category, status, and priority.
type InsightFilter struct {
	filter.Base
	ProjectID *filter.StringField `json:"projectId" filter:"project_id"`
	Category  *filter.StringField `json:"category" filter:"category"`
	Status    *filter.StringField `json:"status" filter:"status"`
	Priority  *filter.StringField `json:"priority" filter:"priority"`
}
