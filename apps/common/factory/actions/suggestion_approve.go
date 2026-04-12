package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// ApproveSuggestionAction approves a suggestion.
type ApproveSuggestionAction struct {
	action.TypedInput[inputs.ApproveSuggestionInput]
	action.RequirePolicy[policies.CanApproveSuggestionPolicy]
}

func (a *ApproveSuggestionAction) Description() string { return "Approve a pending suggestion" }

func (a *ApproveSuggestionAction) Execute(ctx context.Context, actx *action.Context) error {
	input := a.Input(actx)

	_, err := action.Write[entities.Suggestion](actx).Exec(ctx, write.Update{
		ID: actx.EntityID,
		Set: write.Set{
			fields.Suggestion.Status.Value(string(enums.SuggestionStatusApproved)),
			fields.Suggestion.ConvertedTaskID.When(input.PRDID != "").Value(input.PRDID),
		},
	})
	return err
}
