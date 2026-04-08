package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

type InsightApply struct {
	command.Base
}

func (c *InsightApply) Name() string        { return "insight:apply" }
func (c *InsightApply) Description() string { return "Mark an insight as applied" }
func (c *InsightApply) Input() any          { return nil }

func (c *InsightApply) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: insight:apply <id>")
	}
	id := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("Insight")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Insight])

	if _, err := w.Update(ctx).WhereID(id).Set("status", entities.InsightApplied).Exec(ctx); err != nil {
		return fmt.Errorf("apply insight: %w", err)
	}

	cctx.Print("Applied insight %s", id)
	return nil
}
