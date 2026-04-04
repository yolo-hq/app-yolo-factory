package main

import (
	"github.com/uptrace/bun"
	"github.com/yolo-hq/yolo/core/jobs"
	"github.com/yolo-hq/yolo/core/registry"

	bunrepo "github.com/yolo-hq/yolo/core/bun"

	"github.com/yolo-hq/app-yolo-factory/server/factory/actions"
	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/filters"
	factoryjobs "github.com/yolo-hq/app-yolo-factory/server/factory/jobs"
)

// setup registers all entities, repositories, actions, and filters.
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

	// Actions
	registry.RegisterActions("Project",
		&actions.CreateProjectAction{},
		&actions.UpdateProjectAction{},
		&actions.PauseProjectAction{},
		&actions.ResumeProjectAction{},
	)
	registry.RegisterActions("PRD",
		&actions.SubmitPRDAction{},
		&actions.ApprovePRDAction{},
		&actions.ExecutePRDAction{},
	)
	registry.RegisterActions("Task",
		&actions.CancelTaskAction{},
		&actions.RetryTaskAction{},
	)
	registry.RegisterActions("Run",
		&actions.CompleteRunAction{},
	)
	registry.RegisterActions("Question",
		&actions.AnswerQuestionAction{},
	)
	registry.RegisterActions("Suggestion",
		&actions.ApproveSuggestionAction{},
		&actions.RejectSuggestionAction{},
	)

	// Filters
	registry.RegisterFilter("Project", &filters.ProjectFilter{})
	registry.RegisterFilter("PRD", &filters.PRDFilter{})
	registry.RegisterFilter("Task", &filters.TaskFilter{})
	registry.RegisterFilter("Run", &filters.RunFilter{})
	registry.RegisterFilter("Step", &filters.StepFilter{})
	registry.RegisterFilter("Question", &filters.QuestionFilter{})
	registry.RegisterFilter("Suggestion", &filters.SuggestionFilter{})

	// Jobs
	jobs.RegisterHandler(&factoryjobs.PlanPRDJob{})
	jobs.RegisterHandler(&factoryjobs.ExecuteWorkflowJob{})
}
