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

// PRDRow is the projection used for list responses.
type PRDRow struct {
	projection.For[entities.PRD]

	ID             string  `field:"id"`
	Title          string  `field:"title"`
	Status         string  `field:"status"`
	CompletedTasks int     `field:"completedTasks"`
	TotalTasks     int     `field:"totalTasks"`
	TotalCostUSD   float64 `field:"totalCostUsd"`
}

// ListPRDsResponse is the typed response for ListPRDsAction.
type ListPRDsResponse struct {
	PRDs []PRDRow `json:"prds"`
}

// ListPRDsAction lists PRDs with optional filters.
type ListPRDsAction struct {
	action.SkipAllPolicies
	action.TypedInput[inputs.ListPRDsInput]
	action.TypedResponse[ListPRDsResponse]
}

func (a *ListPRDsAction) Description() string { return "List PRDs with optional filters" }

func (a *ListPRDsAction) Execute(ctx context.Context, actx *action.Context) error {
	input := a.Input(actx)

	opts := []read.Option{read.OrderBy("created_at", read.Desc)}
	if input.ProjectID != "" {
		opts = append(opts, read.Eq("project_id", input.ProjectID))
	}
	if input.Status != "" {
		opts = append(opts, read.Eq("status", input.Status))
	}

	prds, err := read.FindMany[PRDRow](ctx, opts...)
	if err != nil {
		return fmt.Errorf("list-prds: %w", err)
	}

	return a.Respond(actx, ListPRDsResponse{PRDs: prds})
}
