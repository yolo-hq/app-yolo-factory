package filters

import (
	"github.com/yolo-hq/yolo/core/filter"
)

// RunFilter filters runs by task, status, agent type, and model.
type RunFilter struct {
	filter.Base
	TaskID    *filter.StringField `json:"task_id" filter:"task_id"`
	Status    *filter.StringField `json:"status" filter:"status"`
	AgentType *filter.StringField `json:"agent_type" filter:"agent_type"`
	Model     *filter.StringField `json:"model" filter:"model"`
}
