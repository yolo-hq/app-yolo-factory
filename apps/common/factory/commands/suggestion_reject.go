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

type SuggestionReject struct {
	command.Base
}

type SuggestionRejectInput struct {
	Reason string `flag:"reason" validate:"required" usage:"Reason for rejection"`
}

func (c *SuggestionReject) Name() string        { return "suggestion:reject" }
func (c *SuggestionReject) Description() string { return "Reject a suggestion" }
func (c *SuggestionReject) Input() any          { return &SuggestionRejectInput{} }

func (c *SuggestionReject) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) == 0 {
		return fmt.Errorf("usage: suggestion:reject <id> --reason <reason>")
	}
	id := cctx.Args[0]

	repo, err := cctx.RepoProvider.Repo("Suggestion")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Suggestion])

	if _, err := w.Update(ctx).WhereID(id).Set(fields.Suggestion.Status.Name(), string(enums.SuggestionStatusRejected)).Exec(ctx); err != nil {
		return fmt.Errorf("reject suggestion: %w", err)
	}

	cctx.Print("Rejected suggestion %s", id)
	return nil
}
