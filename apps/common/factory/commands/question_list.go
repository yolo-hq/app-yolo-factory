package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

type QuestionList struct {
	command.Base
}

type QuestionListInput struct {
	Status string `flag:"status" usage:"Filter by status (open, answered, dismissed)"`
}

func (c *QuestionList) Name() string        { return "question:list" }
func (c *QuestionList) Description() string { return "List questions" }
func (c *QuestionList) Input() any          { return &QuestionListInput{} }

func (c *QuestionList) Execute(ctx context.Context, cctx command.Context) error {
	input, _ := cctx.TypedInput.(*QuestionListInput)

	repo, err := cctx.RepoProvider.Repo("Question")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	r := repo.(entity.ReadRepository[entities.Question])

	opts := entity.FindOptions{
		Sort: &entity.SortParams{Field: "created_at", Order: "desc"},
	}
	if input != nil && input.Status != "" {
		opts.Filters = append(opts.Filters, entity.FilterCondition{
			Field: "status", Operator: entity.OpEq, Value: input.Status,
		})
	}

	result, err := r.FindMany(ctx, opts)
	if err != nil {
		return fmt.Errorf("list questions: %w", err)
	}

	if len(result.Data) == 0 {
		cctx.Print("No questions found.")
		return nil
	}

	tw := cctx.Table()
	fmt.Fprintf(tw, "ID\tSTATUS\tCONFIDENCE\tBODY\n")
	for _, q := range result.Data {
		body := q.Body
		if len(body) > 60 {
			body = body[:57] + "..."
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", q.ID, q.Status, q.Confidence, body)
	}
	tw.Flush()
	return nil
}
