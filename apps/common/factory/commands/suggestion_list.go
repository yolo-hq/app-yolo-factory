package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

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
