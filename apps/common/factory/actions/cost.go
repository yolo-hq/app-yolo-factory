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

// RunCostRow holds cost data from a single run.
type RunCostRow struct {
	projection.For[entities.Run]

	Model   string  `field:"model"`
	CostUSD float64 `field:"costUsd"`
}

// ModelCost aggregates cost and run count for a model.
type ModelCost struct {
	Model   string  `json:"model"`
	CostUSD float64 `json:"costUsd"`
	Runs    int     `json:"runs"`
}

// CostResponse is the typed response for CostAction.
type CostResponse struct {
	Breakdown []ModelCost `json:"breakdown"`
	TotalCost float64     `json:"totalCost"`
	TotalRuns int         `json:"totalRuns"`
}

// CostAction shows a cost breakdown grouped by model.
type CostAction struct {
	action.SkipAllPolicies
	action.TypedInput[inputs.CostInput]
	action.TypedResponse[CostResponse]
}

func (a *CostAction) Description() string { return "Show cost breakdown by model" }

func (a *CostAction) Execute(ctx context.Context, actx *action.Context) error {
	input := a.Input(actx)

	opts := []read.Option{read.OrderBy("created_at", read.Desc)}
	if input.ProjectID != "" {
		opts = append(opts, read.Eq("project_id", input.ProjectID))
	}

	runs, err := read.FindMany[RunCostRow](ctx, opts...)
	if err != nil {
		return fmt.Errorf("cost: %w", err)
	}

	byModel := map[string]*ModelCost{}
	var total float64
	for _, r := range runs {
		mc := byModel[r.Model]
		if mc == nil {
			mc = &ModelCost{Model: r.Model}
			byModel[r.Model] = mc
		}
		mc.CostUSD += r.CostUSD
		mc.Runs++
		total += r.CostUSD
	}

	breakdown := make([]ModelCost, 0, len(byModel))
	for _, mc := range byModel {
		breakdown = append(breakdown, *mc)
	}

	return a.Respond(actx, CostResponse{
		Breakdown: breakdown,
		TotalCost: total,
		TotalRuns: len(runs),
	})
}
