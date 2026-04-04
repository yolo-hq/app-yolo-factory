package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

// --- QuestionList ---

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

// --- QuestionAnswer ---

type QuestionAnswer struct {
	command.Base
}

func (c *QuestionAnswer) Name() string        { return "question:answer" }
func (c *QuestionAnswer) Description() string { return "Answer a question" }

func (c *QuestionAnswer) Execute(ctx context.Context, cctx command.Context) error {
	if len(cctx.Args) < 2 {
		return fmt.Errorf("usage: question:answer <id> <answer text>")
	}
	id := cctx.Args[0]
	answer := strings.Join(cctx.Args[1:], " ")

	repo, err := cctx.RepoProvider.Repo("Question")
	if err != nil {
		return fmt.Errorf("get repo: %w", err)
	}
	w := repo.(entity.WriteRepository[entities.Question])

	now := time.Now()
	if _, err := w.Update(ctx).WhereID(id).
		Set("status", entities.QuestionAnswered).
		Set("answer", answer).
		Set("answered_by", "human").
		Set("answered_at", now).
		Exec(ctx); err != nil {
		return fmt.Errorf("answer question: %w", err)
	}

	cctx.Print("Answered question %s", id)
	return nil
}
