package actions

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/projection"
	"github.com/yolo-hq/yolo/core/read"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

// ProjectDetail is the projection used for single-project responses.
type ProjectDetail struct {
	projection.For[entities.Project]

	ID                string  `field:"id"`
	Name              string  `field:"name"`
	Status            string  `field:"status"`
	RepoURL           string  `field:"repoUrl"`
	LocalPath         string  `field:"localPath"`
	DefaultBranch     string  `field:"defaultBranch"`
	DefaultModel      string  `field:"defaultModel"`
	BudgetMonthlyUSD  float64 `field:"budgetMonthlyUsd"`
	SpentThisMonthUSD float64 `field:"spentThisMonthUsd"`
}

// GetProjectResponse is the typed response for GetProjectAction.
type GetProjectResponse struct {
	Project *ProjectDetail `json:"project"`
}

// GetProjectAction fetches a single project by ID.
type GetProjectAction struct {
	action.SkipAllPolicies
	action.TypedInput[inputs.GetProjectInput]
	action.TypedResponse[GetProjectResponse]
}

func (a *GetProjectAction) Description() string { return "Get a project by ID" }

func (a *GetProjectAction) Execute(ctx context.Context, actx *action.Context) error {
	input := a.Input(actx)

	project, err := read.FindOne[ProjectDetail](ctx, input.ID)
	if err != nil {
		return fmt.Errorf("get-project: %w", err)
	}
	if project.ID == "" {
		return fmt.Errorf("project %s not found", input.ID)
	}

	return a.Respond(actx, GetProjectResponse{Project: &project})
}
