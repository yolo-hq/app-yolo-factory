package filters

import (
	"github.com/yolo-hq/yolo/core/filter"
)

// PRDFilter filters PRDs by project, status, source, and creator.
type PRDFilter struct {
	filter.Base
	ProjectID *filter.StringField `json:"project_id" filter:"project_id"`
	Status    *filter.StringField `json:"status" filter:"status"`
	Source    *filter.StringField `json:"source" filter:"source"`
	CreatedBy *filter.StringField `json:"created_by" filter:"created_by"`
}
