package main

import (
	"github.com/yolo-hq/yolo"
	"github.com/yolo-hq/yolo/core/registry"

	"github.com/yolo-hq/app-yolo-factory/server/factory/actions"
	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/filters"
)

func main() {
	// Entities
	registry.Register(entities.Repo{}, entities.Task{}, entities.Run{}, entities.Question{})

	// Filters
	registry.RegisterFilter("Task", filters.TaskFilter{})
	registry.RegisterFilter("Run", filters.RunFilter{})
	registry.RegisterFilter("Question", filters.QuestionFilter{})

	// Actions
	registry.RegisterActions("Repo",
		&actions.CreateRepoAction{},
		&actions.UpdateRepoAction{},
	)
	registry.RegisterActions("Task",
		&actions.CreateTaskAction{},
		&actions.UpdateTaskAction{},
		&actions.CancelTaskAction{},
	)
	registry.RegisterActions("Run",
		&actions.CreateRunAction{},
		&actions.CompleteRunAction{},
	)
	registry.RegisterActions("Question",
		&actions.CreateQuestionAction{},
		&actions.ResolveQuestionAction{},
	)
	registry.RegisterActions("",
		&actions.ExecuteTaskAction{},
	)

	yolo.MustRunBinary()
}
