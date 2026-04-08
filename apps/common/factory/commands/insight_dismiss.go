package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

type InsightDismiss struct {
	command.Base
}

type InsightDismissInput struct {
	Reason string `flag:"reason" validate:"required" usage:"Reason for dismissal"`
}

func (c *InsightDismiss) Name() string        { return "insight:dismiss" }
func (c *InsightDismiss) Description() string { return "Dismiss an insight" }
func (c *InsightDismiss) Input() any          { return &InsightDismissInput{} }

func (c *InsightDismiss) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: insight:dismiss <id> --reason <reason>")
	}
	id := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("Insight")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Insight])

	if _, err := w.Update(ctx).WhereID(id).Set("status", entities.InsightDismissed).Exec(ctx); err != nil {
		return fmt.Errorf("dismiss insight: %w", err)
	}

	cctx.Print("Dismissed insight %s", id)
	return nil
}
