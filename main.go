package main

import (
	"github.com/yolo-hq/yolo"
	"github.com/yolo-hq/yolo/core/registry"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/filters"

	_ "github.com/yolo-hq/app-yolo-factory/server/factory/actions"
)

func main() {
	registry.Register(entities.Repo{}, entities.Task{}, entities.Run{}, entities.Question{})
	registry.RegisterFilter("Task", filters.TaskFilter{})
	registry.RegisterFilter("Run", filters.RunFilter{})
	registry.RegisterFilter("Question", filters.QuestionFilter{})

	yolo.MustRunBinary()
}
