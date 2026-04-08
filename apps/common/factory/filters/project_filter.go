package filters

import (
	"github.com/yolo-hq/yolo/core/filter"
)

// ProjectFilter filters projects by status and name.
type ProjectFilter struct {
	filter.Base
	Status *filter.StringField `json:"status" filter:"status"`
	Name   *filter.StringField `json:"name" filter:"name"`
}
