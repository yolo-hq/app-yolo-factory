package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
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
	if project.Status != string(enums.ProjectStatusActive) {
		return action.Failure("project must be active to submit a PRD")
	}

	source := input.Source
	if source == "" {
		source = string(enums.PRDSourceManual)
	}

	res, err := action.Write[entities.PRD](actx).Exec(ctx, write.Create{
		FromInput: input,
		Set: write.Set{
			fields.PRD.Status.Value(string(enums.PRDStatusDraft)),
			fields.PRD.CreatedBy.Value("human"),
			fields.PRD.Source.Value(source),
		},
	})
	if err != nil {
		return action.Failure(err.Error())
	}

	actx.Resolve("PRD", res.ID())
	return action.OK()
}
