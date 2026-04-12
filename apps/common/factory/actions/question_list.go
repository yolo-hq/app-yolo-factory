package actions

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/projection"
	"github.com/yolo-hq/yolo/core/read"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

// QuestionRow is the projection used for list responses.
type QuestionRow struct {
	projection.For[entities.Question]

	ID         string `field:"id"`
	Body       string `field:"body"`
	Status     string `field:"status"`
	Confidence string `field:"confidence"`
}

// ListQuestionsResponse is the typed response for ListQuestionsAction.
type ListQuestionsResponse struct {
	Questions []QuestionRow `json:"questions"`
}

// ListQuestionsAction lists questions with optional status filter.
type ListQuestionsAction struct {
	action.SkipAllPolicies
	action.TypedInput[inputs.ListQuestionsInput]
	action.TypedResponse[ListQuestionsResponse]
}

func (a *ListQuestionsAction) Description() string { return "List questions with optional status filter" }

func (a *ListQuestionsAction) Execute(ctx context.Context, actx *action.Context) error {
	input := a.Input(actx)

	opts := []read.Option{read.OrderBy("created_at", read.Desc)}
	if input.Status != "" {
		opts = append(opts, read.Eq("status", input.Status))
	}

	questions, err := read.FindMany[QuestionRow](ctx, opts...)
	if err != nil {
		return fmt.Errorf("list-questions: %w", err)
	}

	return a.Respond(actx, ListQuestionsResponse{Questions: questions})
}
