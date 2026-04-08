package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

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
