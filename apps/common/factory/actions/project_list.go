package actions

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/projection"
	"github.com/yolo-hq/yolo/core/read"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// ProjectRow is the projection used for list responses.
type ProjectRow struct {
	projection.For[entities.Project]

	ID                string  `field:"id"`
	Name              string  `field:"name"`
	Status            string  `field:"status"`
	DefaultModel      string  `field:"defaultModel"`
	BudgetMonthlyUSD  float64 `field:"budgetMonthlyUsd"`
	SpentThisMonthUSD float64 `field:"spentThisMonthUsd"`
}

// ListProjectsResponse is the typed response for ListProjectsAction.
type ListProjectsResponse struct {
	Projects []ProjectRow `json:"projects"`
}

// ListProjectsAction lists all projects.
type ListProjectsAction struct {
	action.SkipAllPolicies
	action.NoInput
	action.TypedResponse[ListProjectsResponse]
}

func (a *ListProjectsAction) Description() string { return "List all projects" }

func (a *ListProjectsAction) Execute(ctx context.Context, actx *action.Context) error {
	projects, err := read.FindMany[ProjectRow](ctx, read.OrderBy("name", read.Asc))
	if err != nil {
		return fmt.Errorf("list-projects: %w", err)
	}

	return a.Respond(actx, ListProjectsResponse{Projects: projects})
}
