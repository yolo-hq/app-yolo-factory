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

type InsightAcknowledge struct {
	command.Base
}

func (c *InsightAcknowledge) Name() string        { return "insight:acknowledge" }
func (c *InsightAcknowledge) Description() string { return "Acknowledge an insight" }
func (c *InsightAcknowledge) Input() any          { return nil }

func (c *InsightAcknowledge) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: insight:acknowledge <id>")
	}
	id := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("Insight")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Insight])

	if _, err := w.Update(ctx).WhereID(id).Set(fields.Insight.Status.Name(), string(enums.InsightStatusAcknowledged)).Exec(ctx); err != nil {
		return fmt.Errorf("acknowledge insight: %w", err)
	}

	cctx.Print("Acknowledged insight %s", id)
	return nil
}
