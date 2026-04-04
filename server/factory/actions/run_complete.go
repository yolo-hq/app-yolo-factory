package actions

import (
	"context"
	"time"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/inputs"
)

// CompleteRunAction records run completion metrics.
type CompleteRunAction struct {
	action.TypedInput[inputs.CompleteRunInput]
}

func (a *CompleteRunAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	_, r := action.FindOrFail[entities.Run](ctx, action.ReadRepo[entities.Run](actx), actx.EntityID)
	if r != nil {
		return *r
	}

	input := a.Input(actx)
	now := time.Now()

	_, err := action.Write[entities.Run](actx).Exec(ctx, write.Update{
		ID: actx.EntityID,
		Set: write.Set{
			write.NewField[string]("status").Value(input.Status),
			write.NewField[float64]("cost_usd").Value(input.CostUSD),
			write.NewField[int]("tokens_in").Value(input.TokensIn),
			write.NewField[int]("tokens_out").Value(input.TokensOut),
			write.NewField[int]("duration_ms").Value(input.DurationMS),
			write.NewField[int]("num_turns").Value(input.NumTurns),
			write.NewField[string]("error").Value(input.Error),
			write.NewField[string]("commit_hash").Value(input.CommitHash),
			write.NewField[string]("files_changed").Value(input.FilesChanged),
			write.NewField[string]("result").Value(input.Result),
			write.NewField[string]("session_id").Value(input.SessionID),
			write.NewField[*time.Time]("completed_at").Value(&now),
		},
	})
	if err != nil {
		return action.Failure(err.Error())
	}

	actx.Resolve("Run", actx.EntityID)
	return action.OK()
}
