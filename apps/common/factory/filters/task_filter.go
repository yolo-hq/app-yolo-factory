package filters

import (
	"github.com/yolo-hq/yolo/core/filter"
)

// TaskFilter filters tasks by PRD, project, status, branch, and sequence.
type TaskFilter struct {
	filter.Base
	PRDID     *filter.StringField      `json:"prdId" filter:"prd_id"`
	ProjectID *filter.StringField      `json:"projectId" filter:"project_id"`
	Status    *filter.StringField      `json:"status" filter:"status"`
	Branch    *filter.StringField      `json:"branch" filter:"branch"`
	Sequence  *filter.NumberField[int] `json:"sequence" filter:"sequence"`
}
