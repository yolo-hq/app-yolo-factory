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

// SubmitPRDAction creates a new PRD for a project.
// The input's resolves:"Project" tag promotes ProjectID as the primary entity
// so that CanSubmitPRDPolicy can load and check the project status.
type SubmitPRDAction struct {
	action.TypedInput[inputs.SubmitPRDInput]
	action.RequirePolicy[policies.CanSubmitPRDPolicy]
}

func (a *SubmitPRDAction) Description() string { return "Submit a new PRD for a project" }

func (a *SubmitPRDAction) Execute(ctx context.Context, actx *action.Context) error {
	input := a.Input(actx)

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
		return err
	}

	actx.Resolve("PRD", res.ID())
	return nil
}
