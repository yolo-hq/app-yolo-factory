package main

import (
	"github.com/uptrace/bun"
	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/registry"

	bunrepo "github.com/yolo-hq/yolo/core/bun"

	"github.com/yolo-hq/app-yolo-factory/server/factory/actions"
	"github.com/yolo-hq/app-yolo-factory/server/factory/commands"
	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/filters"
)

// setup registers all entities, repositories, filters, actions, and commands.
// Called by both main() and test code.
func setup() {
	// Entities
	registry.Register(entities.Repo{}, entities.Run{}, entities.Task{}, entities.Question{})

	// Repositories
	registry.RegisterRepoFactory("Repo", func(db any) (any, any) {
		return bunrepo.NewReadRepository[entities.Repo](db.(*bun.DB)), bunrepo.NewWriteRepository[entities.Repo](db.(*bun.DB))
	})
	registry.RegisterRepoFactory("Run", func(db any) (any, any) {
		return bunrepo.NewReadRepository[entities.Run](db.(*bun.DB)), bunrepo.NewWriteRepository[entities.Run](db.(*bun.DB))
	})
	registry.RegisterRepoFactory("Task", func(db any) (any, any) {
		return bunrepo.NewReadRepository[entities.Task](db.(*bun.DB)), bunrepo.NewWriteRepository[entities.Task](db.(*bun.DB))
	})
	registry.RegisterRepoFactory("Question", func(db any) (any, any) {
		return bunrepo.NewReadRepository[entities.Question](db.(*bun.DB)), bunrepo.NewWriteRepository[entities.Question](db.(*bun.DB))
	})

	// Filters
	registry.RegisterFilter("Task", filters.TaskFilter{})
	registry.RegisterFilter("Question", filters.QuestionFilter{})
	registry.RegisterFilter("Run", filters.RunFilter{})

	// Actions
	registry.RegisterActions("Task", &actions.CreateTaskAction{}, &actions.CancelTaskAction{}, &actions.ExecuteTaskAction{}, &actions.UpdateTaskAction{})
	registry.RegisterActions("Question", &actions.CreateQuestionAction{}, &actions.ResolveQuestionAction{})
	registry.RegisterActions("Repo", &actions.CreateRepoAction{}, &actions.UpdateRepoAction{})
	registry.RegisterActions("Run", &actions.CompleteRunAction{}, &actions.CreateRunAction{})

	// Commands
	command.Register(&commands.RetryFailed{})
	command.Register(&commands.CleanupRuns{})
}
