package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// --- SuggestionList ---

type SuggestionList struct {
	command.Base
}

type SuggestionListInput struct {
	Project  string `flag:"project" usage:"Filter by project ID"`
	Category string `flag:"category" usage:"Filter by category"`
	Priority string `flag:"priority" usage:"Filter by priority"`
	Status   string `flag:"status" usage:"Filter by status"`
}

func (c *SuggestionList) Name() string        { return "suggestion:list" }
func (c *SuggestionList) Description() string { return "List suggestions" }
func (c *SuggestionList) Input() any          { return &SuggestionListInput{} }

func (c *SuggestionList) Execute(ctx context.Context, cctx command.Context) error {
	input, _ := cctx.TypedInput.(*SuggestionListInput)

	repo, err := cctx.RepoProvider.Repo("Suggestion")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	r := repo.(entity.ReadRepository[entities.Suggestion])

	opts := entity.FindOptions{
		Sort: &entity.SortParams{Field: "created_at", Order: "desc"},
	}
	if input != nil {
		if input.Project != "" {
			opts.Filters = append(opts.Filters, entity.FilterCondition{
				Field: "project_id", Operator: entity.OpEq, Value: input.Project,
			})
		}
		if input.Category != "" {
			opts.Filters = append(opts.Filters, entity.FilterCondition{
				Field: "category", Operator: entity.OpEq, Value: input.Category,
			})
		}
		if input.Priority != "" {
			opts.Filters = append(opts.Filters, entity.FilterCondition{
				Field: "priority", Operator: entity.OpEq, Value: input.Priority,
			})
		}
		if input.Status != "" {
			opts.Filters = append(opts.Filters, entity.FilterCondition{
				Field: "status", Operator: entity.OpEq, Value: input.Status,
			})
		}
	}

	result, err := r.FindMany(ctx, opts)
	if err != nil {
		return fmt.Errorf("list suggestions: %w", err)
	}

	if len(result.Data) == 0 {
		cctx.Print("No suggestions found.")
		return nil
	}

	tw := cctx.Table()
	fmt.Fprintf(tw, "ID\tTITLE\tSOURCE\tCATEGORY\tPRIORITY\tSTATUS\n")
	for _, s := range result.Data {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			s.ID, s.Title, s.Source, s.Category, s.Priority, s.Status)
	}
	tw.Flush()
	return nil
}

// --- SuggestionApprove ---

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

	if _, err := w.Update(ctx).WhereID(id).Set("status", entities.SuggestionApproved).Exec(ctx); err != nil {
		return fmt.Errorf("approve suggestion: %w", err)
	}

	cctx.Print("Approved suggestion %s", id)
	return nil
}

// --- SuggestionReject ---

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

	if _, err := w.Update(ctx).WhereID(id).Set("status", entities.SuggestionRejected).Exec(ctx); err != nil {
		return fmt.Errorf("reject suggestion: %w", err)
	}

	cctx.Print("Rejected suggestion %s", id)
	return nil
}
