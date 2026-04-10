package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

type ProjectPause struct {
	command.Base
}

func (c *ProjectPause) Name() string        { return "project:pause" }
func (c *ProjectPause) Description() string { return "Pause a project" }

func (c *ProjectPause) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: project:pause <id>")
	}
	id := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("Project")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Project])

	if _, err := w.Update(ctx).WhereID(id).Set(fields.Project.Status.Name(), string(enums.ProjectStatusPaused)).Exec(ctx); err != nil {
		return fmt.Errorf("pause project: %w", err)
	}

	cctx.Print("Paused project %s", id)
	return nil
}
