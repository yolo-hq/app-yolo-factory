package filters

import (
	"github.com/yolo-hq/yolo/core/filter"
)

// SuggestionFilter filters suggestions by project, source, category, status, and priority.
type SuggestionFilter struct {
	filter.Base
	ProjectID *filter.StringField `json:"projectId" filter:"project_id"`
	Source    *filter.StringField `json:"source" filter:"source"`
	Category  *filter.StringField `json:"category" filter:"category"`
	Status    *filter.StringField `json:"status" filter:"status"`
	Priority  *filter.StringField `json:"priority" filter:"priority"`
}
