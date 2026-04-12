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

// SuggestionRow is the projection used for list responses.
type SuggestionRow struct {
	projection.For[entities.Suggestion]

	ID       string `field:"id"`
	Title    string `field:"title"`
	Source   string `field:"source"`
	Category string `field:"category"`
	Priority string `field:"priority"`
	Status   string `field:"status"`
}

// ListSuggestionsResponse is the typed response for ListSuggestionsAction.
type ListSuggestionsResponse struct {
	Suggestions []SuggestionRow `json:"suggestions"`
}

// ListSuggestionsAction lists suggestions with optional filters.
type ListSuggestionsAction struct {
	action.SkipAllPolicies
	action.TypedInput[inputs.ListSuggestionsInput]
	action.TypedResponse[ListSuggestionsResponse]
}

func (a *ListSuggestionsAction) Description() string {
	return "List suggestions with optional filters"
}

func (a *ListSuggestionsAction) Execute(ctx context.Context, actx *action.Context) error {
	input := a.Input(actx)

	opts := []read.Option{read.OrderBy("created_at", read.Desc)}
	if input.ProjectID != "" {
		opts = append(opts, read.Eq("project_id", input.ProjectID))
	}
	if input.Category != "" {
		opts = append(opts, read.Eq("category", input.Category))
	}
	if input.Priority != "" {
		opts = append(opts, read.Eq("priority", input.Priority))
	}
	if input.Status != "" {
		opts = append(opts, read.Eq("status", input.Status))
	}

	suggestions, err := read.FindMany[SuggestionRow](ctx, opts...)
	if err != nil {
		return fmt.Errorf("list-suggestions: %w", err)
	}

	return a.Respond(actx, ListSuggestionsResponse{Suggestions: suggestions})
}
