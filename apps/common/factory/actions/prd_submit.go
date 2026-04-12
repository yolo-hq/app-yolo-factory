package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/.yolo/repos"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/constants"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/policies"
)

// SubmitPRDAction creates a new PRD for a project.
// The input's resolves:"Project" tag promotes ProjectID as the primary entity
// so that CanSubmitPRDPolicy can load and check the project status.
type SubmitPRDAction struct {
	action.RequirePolicy[policies.CanSubmitPRDPolicy]
	action.TypedInput[inputs.SubmitPRDInput]
}

func (a *SubmitPRDAction) Description() string { return "Submit a new PRD for a project" }

func (a *SubmitPRDAction) Execute(ctx context.Context, actx *action.Context) error {
	input := a.Input(actx)

	source := input.Source
	if source == "" {
		source = string(enums.PRDSourceManual)
	}

	_, err := repos.PRD.CreateFromInput(ctx, actx, input,
		fields.PRD.Status.Value(string(enums.PRDStatusDraft)),
		fields.PRD.CreatedBy.Value(constants.ActorHuman),
		fields.PRD.Source.Value(source),
	)
	return err
}
