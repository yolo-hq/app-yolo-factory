package queries

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/projection"
	"github.com/yolo-hq/yolo/core/query"
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

// CostResponse is the typed response for CostQuery.
type CostResponse struct {
	Breakdown []CostByModel `json:"breakdown"`
	TotalCost float64       `json:"total_cost"`
	TotalRuns int           `json:"total_runs"`
}

// CostQuery shows a cost breakdown grouped by model.
type CostQuery struct {
	query.Base
	query.TypedInput[inputs.CostInput]
	query.Returns[CostResponse]
}

func (q *CostQuery) Description() string { return "Cost breakdown by model and period" }

func (q *CostQuery) Execute(ctx context.Context, qctx *query.Context) error {
	input := q.Input(qctx)

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

	return q.Respond(qctx, CostResponse{
		Breakdown: breakdown,
		TotalCost: total,
		TotalRuns: totalRuns,
	})
}
