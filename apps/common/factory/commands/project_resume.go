package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

type ProjectResume struct {
	command.Base
}

func (c *ProjectResume) Name() string        { return "project:resume" }
func (c *ProjectResume) Description() string { return "Resume a paused project" }

func (c *ProjectResume) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: project:resume <id>")
	}
	id := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("Project")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Project])

	if _, err := w.Update(ctx).WhereID(id).Set(fields.Project.Status.Name(), entities.ProjectActive).Exec(ctx); err != nil {
		return fmt.Errorf("resume project: %w", err)
	}

	cctx.Print("Resumed project %s", id)
	return nil
}
