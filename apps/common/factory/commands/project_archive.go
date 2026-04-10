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

type ProjectArchive struct {
	command.Base
}

func (c *ProjectArchive) Name() string        { return "project:archive" }
func (c *ProjectArchive) Description() string { return "Archive a project" }

func (c *ProjectArchive) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: project:archive <id>")
	}
	id := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("Project")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Project])

	if _, err := w.Update(ctx).WhereID(id).Set(fields.Project.Status.Name(), string(enums.ProjectStatusArchived)).Exec(ctx); err != nil {
		return fmt.Errorf("archive project: %w", err)
	}

	cctx.Print("Archived project %s", id)
	return nil
}
