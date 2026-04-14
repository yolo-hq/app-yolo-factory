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

// CostResponse is the typed response for Cost.
type CostResponse struct {
	Breakdown []CostByModel `json:"breakdown"`
	TotalCost float64       `json:"totalCost"`
	TotalRuns int           `json:"totalRuns"`
}

// Cost shows a cost breakdown grouped by model.
func Cost(ctx context.Context, qctx *query.Context, in inputs.CostInput) (CostResponse, error) {
	opts := []read.Option{}
	if in.ProjectID != "" {
		opts = append(opts, read.Eq("project_id", in.ProjectID))
	}

	breakdown, err := read.FindMany[CostByModel](ctx, opts...)
	if err != nil {
		return CostResponse{}, fmt.Errorf("cost: %w", err)
	}

	var total float64
	var totalRuns int
	for _, b := range breakdown {
		total += b.Total
		totalRuns += b.Runs
	}

	return CostResponse{
		Breakdown: breakdown,
		TotalCost: total,
		TotalRuns: totalRuns,
	}, nil
}
