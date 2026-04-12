package actions

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/.yolo/repos"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// RejectSuggestionAction rejects a suggestion with a reason.
type RejectSuggestionAction struct {
	action.RequirePolicy[policies.CanRejectSuggestionPolicy]
	action.TypedInput[inputs.RejectSuggestionInput]
}

func (a *RejectSuggestionAction) Description() string { return "Reject a pending suggestion" }

func (a *RejectSuggestionAction) Execute(ctx context.Context, actx *action.Context) error {
	// input consumed for validation; reason not stored on entity.
	_ = a.Input(actx)

	_, err := repos.Suggestion.UpdateEntity(ctx, actx, write.Set{
		fields.Suggestion.Status.Value(string(enums.SuggestionStatusRejected)),
	})
	if err != nil {
		return fmt.Errorf("reject-suggestion: %w", err)
	}
	return nil
}
