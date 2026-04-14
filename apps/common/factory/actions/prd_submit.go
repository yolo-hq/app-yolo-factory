package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/action"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/.yolo/repos"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/constants"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

// PRDSubmit creates a new PRD for a project. The input's resolves:"Project"
// tag promotes ProjectID as the primary entity so CanSubmitPRDPolicy can
// load and check the project status.
//
// @policy CanSubmitPRDPolicy
func PRDSubmit(ctx context.Context, actx *action.Context, in inputs.SubmitPRDInput) error {
	source := in.Source
	if source == "" {
		source = string(enums.PRDSourceManual)
	}

	result, err := repos.PRD.CreateFromInput(ctx, actx, in,
		fields.PRD.Status.Value(string(enums.PRDStatusDraft)),
		fields.PRD.CreatedBy.Value(constants.ActorHuman),
		fields.PRD.Source.Value(source),
	)
	if err != nil {
		return err
	}

	actx.Resolve("PRD", result.ID())
	return nil
}
