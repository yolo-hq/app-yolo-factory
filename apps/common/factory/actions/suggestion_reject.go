package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

// RejectSuggestionAction rejects a suggestion with a reason.
type RejectSuggestionAction struct {
	action.TypedInput[inputs.RejectSuggestionInput]
}

func (a *RejectSuggestionAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	_, r := action.FindOrFail[entities.Suggestion](ctx, action.ReadRepo[entities.Suggestion](actx), actx.EntityID)
	if r != nil {
		return *r
	}

	// input consumed for validation; reason not stored on entity (no field for it).
	_ = a.Input(actx)

	_, err := action.Write[entities.Suggestion](actx).Exec(ctx, write.Update{
		ID:  actx.EntityID,
		Set: write.Set{write.NewField[string]("status").Value(entities.SuggestionRejected)},
	})
	if err != nil {
		return action.Failure(err.Error())
	}

	actx.Resolve("Suggestion", actx.EntityID)
	return action.OK()
}
