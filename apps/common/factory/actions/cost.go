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

// CostByModel groups run cost by model using a single GROUP BY query.
type CostByModel struct {
	projection.For[entities.Run]
	Model string  `field:"model" group:"true"`
	Total float64 `aggregate:"sum:cost_usd"`
	Runs  int     `aggregate:"count"`
}

// costResponse is the typed response for CostAction.
type costResponse struct {
	Breakdown []CostByModel `json:"breakdown"`
	TotalCost float64       `json:"totalCost"`
	TotalRuns int           `json:"totalRuns"`
}

// CostAction shows a cost breakdown grouped by model.
type CostAction struct {
	action.TypedInput[inputs.CostInput]
	action.TypedResponse[costResponse]
	action.SkipAllPolicies
}

func (a *CostAction) ReadOnly() bool      { return true }
func (a *CostAction) Description() string { return "Cost breakdown by model and period" }

func (a *CostAction) Execute(ctx context.Context, actx *action.Context) error {
	input := a.Input(actx)

	opts := []read.Option{}
	if input.ProjectID != "" {
		opts = append(opts, read.Eq("project_id", input.ProjectID))
	}

	breakdown, err := read.FindMany[CostByModel](ctx, opts...)
	if err != nil {
		return fmt.Errorf("cost: %w", err)
	}

	var total float64
	var totalRuns int
	for _, b := range breakdown {
		total += b.Total
		totalRuns += b.Runs
	}

	return a.Respond(actx, costResponse{
		Breakdown: breakdown,
		TotalCost: total,
		TotalRuns: totalRuns,
	})
}
