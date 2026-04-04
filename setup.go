package main

import (
	"github.com/uptrace/bun"
	"github.com/yolo-hq/yolo/core/registry"

	bunrepo "github.com/yolo-hq/yolo/core/bun"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

// setup registers all entities and repositories.
// Actions, filters, and commands will be added in issue #23.
func setup() {
	// Entities
	registry.Register(
		entities.Project{},
		entities.PRD{},
		entities.Task{},
		entities.Run{},
		entities.Step{},
		entities.Review{},
		entities.Question{},
		entities.Suggestion{},
	)

	// Repositories
	registry.RegisterRepoFactory("Project", func(db any) (any, any) {
		return bunrepo.NewReadRepository[entities.Project](db.(*bun.DB)), bunrepo.NewWriteRepository[entities.Project](db.(*bun.DB))
	})
	registry.RegisterRepoFactory("PRD", func(db any) (any, any) {
		return bunrepo.NewReadRepository[entities.PRD](db.(*bun.DB)), bunrepo.NewWriteRepository[entities.PRD](db.(*bun.DB))
	})
	registry.RegisterRepoFactory("Task", func(db any) (any, any) {
		return bunrepo.NewReadRepository[entities.Task](db.(*bun.DB)), bunrepo.NewWriteRepository[entities.Task](db.(*bun.DB))
	})
	registry.RegisterRepoFactory("Run", func(db any) (any, any) {
		return bunrepo.NewReadRepository[entities.Run](db.(*bun.DB)), bunrepo.NewWriteRepository[entities.Run](db.(*bun.DB))
	})
	registry.RegisterRepoFactory("Step", func(db any) (any, any) {
		return bunrepo.NewReadRepository[entities.Step](db.(*bun.DB)), bunrepo.NewWriteRepository[entities.Step](db.(*bun.DB))
	})
	registry.RegisterRepoFactory("Review", func(db any) (any, any) {
		return bunrepo.NewReadRepository[entities.Review](db.(*bun.DB)), bunrepo.NewWriteRepository[entities.Review](db.(*bun.DB))
	})
	registry.RegisterRepoFactory("Question", func(db any) (any, any) {
		return bunrepo.NewReadRepository[entities.Question](db.(*bun.DB)), bunrepo.NewWriteRepository[entities.Question](db.(*bun.DB))
	})
	registry.RegisterRepoFactory("Suggestion", func(db any) (any, any) {
		return bunrepo.NewReadRepository[entities.Suggestion](db.(*bun.DB)), bunrepo.NewWriteRepository[entities.Suggestion](db.(*bun.DB))
	})
}
