package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

type ProjectAdd struct {
	command.Base
}

type ProjectAddInput struct {
	Name      string `flag:"name" validate:"required" usage:"Project name"`
	RepoURL   string `flag:"repo" validate:"required" usage:"Git repository URL"`
	LocalPath string `flag:"path" validate:"required" usage:"Local clone path"`
	Branch    string `flag:"branch" usage:"Default branch"`
	Model     string `flag:"model" usage:"Default model"`
}

func (c *ProjectAdd) Name() string        { return "project:add" }
func (c *ProjectAdd) Description() string { return "Add a new project" }
func (c *ProjectAdd) Input() any          { return &ProjectAddInput{} }

func (c *ProjectAdd) Execute(ctx context.Context, cctx command.Context) error {
	input, _ := cctx.TypedInput.(*ProjectAddInput)
	repo, err := cctx.RepoProvider.Repo("Project")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Project])

	p := &entities.Project{
		Name:          input.Name,
		RepoURL:       input.RepoURL,
		LocalPath:     input.LocalPath,
		DefaultBranch: input.Branch,
		DefaultModel:  input.Model,
		Status:        string(enums.ProjectStatusActive),
	}
	if p.DefaultBranch == "" {
		p.DefaultBranch = "main"
	}
	if p.DefaultModel == "" {
		p.DefaultModel = "sonnet"
	}

	created, err := w.Insert(ctx, p)
	if err != nil {
		return fmt.Errorf("insert project: %w", err)
	}

	cctx.Print("Created project %s (%s)", created.Name, created.ID)
	return nil
}
