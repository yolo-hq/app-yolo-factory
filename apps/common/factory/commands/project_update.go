package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

type ProjectUpdate struct {
	command.Base
}

type ProjectUpdateInput struct {
	Name    string  `flag:"name" usage:"Project name"`
	Branch  string  `flag:"branch" usage:"Default branch"`
	Model   string  `flag:"model" usage:"Default model"`
	Budget  float64 `flag:"budget" usage:"Monthly budget USD"`
	Retries int     `flag:"retries" usage:"Max retries"`
}

func (c *ProjectUpdate) Name() string        { return "project:update" }
func (c *ProjectUpdate) Description() string { return "Update a project" }
func (c *ProjectUpdate) Input() any          { return &ProjectUpdateInput{} }

func (c *ProjectUpdate) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: project:update <id> [flags]")
	}
	id := cctx.Args[0]
	input, _ := cctx.TypedInput.(*ProjectUpdateInput)

	repo, err := cctx.RepoProvider.Repo("Project")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Project])

	ub := w.Update(ctx).WhereID(id)
	if input.Name != "" {
		ub = ub.Set("name", input.Name)
	}
	if input.Branch != "" {
		ub = ub.Set("default_branch", input.Branch)
	}
	if input.Model != "" {
		ub = ub.Set("default_model", input.Model)
	}
	if input.Budget > 0 {
		ub = ub.Set("budget_monthly_usd", input.Budget)
	}
	if input.Retries > 0 {
		ub = ub.Set("max_retries", input.Retries)
	}

	if _, err := ub.Exec(ctx); err != nil {
		return fmt.Errorf("update project: %w", err)
	}

	cctx.Print("Updated project %s", id)
	return nil
}
