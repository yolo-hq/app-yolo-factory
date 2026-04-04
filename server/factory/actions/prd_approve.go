package actions

import (
	"context"
	"time"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

// ApprovePRDAction approves a draft PRD.
type ApprovePRDAction struct {
	action.NoInput
}

func (a *ApprovePRDAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	prd, r := action.FindOrFail[entities.PRD](ctx, action.ReadRepo[entities.PRD](actx), actx.EntityID)
	if r != nil {
		return *r
	}

	if prd.Status != entities.PRDDraft {
		return action.Failure("PRD must be in draft status to approve")
	}

	now := time.Now()
	_, err := action.Write[entities.PRD](actx).Exec(ctx, write.Update{
		ID: actx.EntityID,
		Set: write.Set{
			write.NewField[string]("status").Value(entities.PRDApproved),
			write.NewField[*time.Time]("approved_at").Value(&now),
		},
	})
	if err != nil {
		return action.Failure(err.Error())
	}

	actx.Resolve("PRD", actx.EntityID)
	return action.OK()
}
