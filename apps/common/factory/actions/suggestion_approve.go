package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

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

func (a *ApproveSuggestionAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	input := a.Input(actx)

	set := write.Set{
		write.NewField[string]("status").Value(entities.SuggestionApproved),
	}
	if input.PRDID != "" {
		set = append(set, write.NewField[string]("converted_task_id").Value(input.PRDID))
	}

	if r := action.ExecUpdate[entities.Suggestion](ctx, actx, set); r != nil {
		return *r
	}
	actx.Resolve("Suggestion", actx.EntityID)
	return action.OK()
}
