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

type PRDApprove struct {
	command.Base
}

func (c *PRDApprove) Name() string        { return "prd:approve" }
func (c *PRDApprove) Description() string { return "Approve a PRD" }

func (c *PRDApprove) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: prd:approve <id>")
	}
	id := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("PRD")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.PRD])

	if _, err := w.Update(ctx).WhereID(id).Set(fields.PRD.Status.Name(), string(enums.PRDStatusApproved)).Exec(ctx); err != nil {
		return fmt.Errorf("approve prd: %w", err)
	}

	cctx.Print("Approved PRD %s", id)
	return nil
}
