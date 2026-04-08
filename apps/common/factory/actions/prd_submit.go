package actions

import (
	"context"

	yolocontext "github.com/yolo-hq/yolo/core/context"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

// SubmitPRDAction creates a new PRD for a project.
type SubmitPRDAction struct {
	action.TypedInput[inputs.SubmitPRDInput]
	action.PublicAccess
}

func (a *SubmitPRDAction) Description() string { return "Submit a new PRD for a project" }

func (a *SubmitPRDAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	input := a.Input(actx)

	// Validate project exists and is active.
	project, r := action.FindOrFail[entities.Project](ctx, action.ReadRepo[entities.Project](actx), input.ProjectID)
	if r != nil {
		return *r
	}
	if project.Status != entities.ProjectActive {
		return action.Failure("project must be active to submit a PRD")
	}

	source := input.Source
	if source == "" {
		source = entities.SourceManual
	}

	res, err := action.Write[entities.PRD](actx).Exec(ctx, write.Create{
		FromInput: input,
		Set: write.Set{
			write.NewField[string]("status").Value(entities.PRDDraft),
			write.NewField[string]("created_by").Value("human"),
			write.NewField[string]("source").Value(source),
		},
	})
	if err != nil {
		return action.Failure(err.Error())
	}

	yolocontext.TrackWrite(ctx, "", "PRD", res.ID(), yolocontext.OpCreate)
	actx.Resolve("PRD", res.ID())
	return action.OK()
}
