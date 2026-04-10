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

type SuggestionApprove struct {
	command.Base
}

type SuggestionApproveInput struct {
	PRD string `flag:"prd" usage:"Optional PRD to associate"`
}

func (c *SuggestionApprove) Name() string        { return "suggestion:approve" }
func (c *SuggestionApprove) Description() string { return "Approve a suggestion" }
func (c *SuggestionApprove) Input() any          { return &SuggestionApproveInput{} }

func (c *SuggestionApprove) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: suggestion:approve <id> [--prd]")
	}
	id := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("Suggestion")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Suggestion])

	if _, err := w.Update(ctx).WhereID(id).Set(fields.Suggestion.Status.Name(), string(enums.SuggestionStatusApproved)).Exec(ctx); err != nil {
		return fmt.Errorf("approve suggestion: %w", err)
	}

	cctx.Print("Approved suggestion %s", id)
	return nil
}
