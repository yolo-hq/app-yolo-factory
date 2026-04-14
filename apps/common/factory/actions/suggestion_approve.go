package actions

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// ApproveSuggestionAction approves a suggestion.
type ApproveSuggestionAction struct {
	action.RequirePolicy[policies.CanApproveSuggestionPolicy]
	action.TypedInput[inputs.ApproveSuggestionInput]
}

func (a *ApproveSuggestionAction) Description() string { return "Approve a pending suggestion" }

func (a *ApproveSuggestionAction) Execute(ctx context.Context, actx *action.Context) error {
	input := a.Input(actx)

	_, err := sm.Suggestion.Approve(ctx, actx, actx.EntityID, write.Set{
		fields.Suggestion.ConvertedTaskID.When(input.PRDID != "").Value(input.PRDID),
	})
	if err != nil {
		return fmt.Errorf("approve-suggestion: %w", err)
	}
	return nil
}
