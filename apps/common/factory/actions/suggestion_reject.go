package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// RejectSuggestionAction rejects a suggestion with a reason.
type RejectSuggestionAction struct {
	action.TypedInput[inputs.RejectSuggestionInput]
	action.RequirePolicy[policies.CanRejectSuggestionPolicy]
}

func (a *RejectSuggestionAction) Description() string { return "Reject a pending suggestion" }

func (a *RejectSuggestionAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	// input consumed for validation; reason not stored on entity.
	_ = a.Input(actx)

	return action.ExecUpdate[entities.Suggestion](ctx, actx, write.Set{
		write.NewField[string]("status").Value(entities.SuggestionRejected),
	})
}
