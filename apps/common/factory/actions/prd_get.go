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

// PRDDetail is the projection used for single-PRD responses.
type PRDDetail struct {
	projection.For[entities.PRD]

	ID                 string  `field:"id"`
	Title              string  `field:"title"`
	Status             string  `field:"status"`
	Source             string  `field:"source"`
	Body               string  `field:"body"`
	AcceptanceCriteria string  `field:"acceptanceCriteria"`
	CompletedTasks     int     `field:"completedTasks"`
	TotalTasks         int     `field:"totalTasks"`
	TotalCostUSD       float64 `field:"totalCostUsd"`
}

// GetPRDResponse is the typed response for GetPRDAction.
type GetPRDResponse struct {
	PRD *PRDDetail `json:"prd"`
}

// GetPRDAction fetches a single PRD by ID.
type GetPRDAction struct {
	action.SkipAllPolicies
	action.TypedInput[inputs.GetPRDInput]
	action.TypedResponse[GetPRDResponse]
}

func (a *GetPRDAction) Description() string { return "Get a PRD by ID" }

func (a *GetPRDAction) Execute(ctx context.Context, actx *action.Context) error {
	input := a.Input(actx)

	prd, err := read.FindOne[PRDDetail](ctx, input.ID)
	if err != nil {
		return fmt.Errorf("get-prd: %w", err)
	}
	if prd.ID == "" {
		return fmt.Errorf("PRD %s not found", input.ID)
	}

	return a.Respond(actx, GetPRDResponse{PRD: &prd})
}
