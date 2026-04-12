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

// costRunRow holds cost data from a single run.
type costRunRow struct {
	projection.For[entities.Run]

	Model   string  `field:"model"`
	CostUSD float64 `field:"costUsd"`
}

// modelCost aggregates cost and run count for a model.
type modelCost struct {
	Model   string  `json:"model"`
	CostUSD float64 `json:"costUsd"`
	Runs    int     `json:"runs"`
}

// costResponse is the typed response for CostAction.
type costResponse struct {
	Breakdown []modelCost `json:"breakdown"`
	TotalCost float64     `json:"totalCost"`
	TotalRuns int         `json:"totalRuns"`
}

// CostAction shows a cost breakdown grouped by model.
type CostAction struct {
	action.TypedInput[inputs.CostInput]
	action.TypedResponse[costResponse]
	action.SkipAllPolicies
}

func (a *CostAction) ReadOnly() bool     { return true }
func (a *CostAction) Description() string { return "Cost breakdown by model and period" }

func (a *CostAction) Execute(ctx context.Context, actx *action.Context) error {
	input := a.Input(actx)

	opts := []read.Option{read.OrderBy("created_at", read.Desc)}
	if input.ProjectID != "" {
		opts = append(opts, read.Eq("project_id", input.ProjectID))
	}

	runs, err := read.FindMany[costRunRow](ctx, opts...)
	if err != nil {
		return fmt.Errorf("cost: %w", err)
	}

	byModel := map[string]*modelCost{}
	var total float64
	for _, r := range runs {
		mc := byModel[r.Model]
		if mc == nil {
			mc = &modelCost{Model: r.Model}
			byModel[r.Model] = mc
		}
		mc.CostUSD += r.CostUSD
		mc.Runs++
		total += r.CostUSD
	}

	breakdown := make([]modelCost, 0, len(byModel))
	for _, mc := range byModel {
		breakdown = append(breakdown, *mc)
	}

	return a.Respond(actx, costResponse{
		Breakdown: breakdown,
		TotalCost: total,
		TotalRuns: len(runs),
	})
}
