package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/inputs"
)

// ApproveSuggestionAction approves a suggestion.
type ApproveSuggestionAction struct {
	action.TypedInput[inputs.ApproveSuggestionInput]
}

func (a *ApproveSuggestionAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	_, r := action.FindOrFail[entities.Suggestion](ctx, action.ReadRepo[entities.Suggestion](actx), actx.EntityID)
	if r != nil {
		return *r
	}

	input := a.Input(actx)

	set := write.Set{
		write.NewField[string]("status").Value("approved"),
	}
	if input.PRDID != "" {
		set = append(set, write.NewField[string]("converted_task_id").Value(input.PRDID))
	}

	_, err := action.Write[entities.Suggestion](actx).Exec(ctx, write.Update{
		ID:  actx.EntityID,
		Set: set,
	})
	if err != nil {
		return action.Failure(err.Error())
	}

	actx.Resolve("Suggestion", actx.EntityID)
	return action.OK()
}
