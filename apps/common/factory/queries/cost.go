package queries

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/projection"
	"github.com/yolo-hq/yolo/core/query"
	"github.com/yolo-hq/yolo/core/read"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
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

// CostQuery shows a cost breakdown grouped by model.
type CostQuery struct {
	query.Base

	ProjectID string `arg:"projectId"`
}

func (q *CostQuery) Execute(ctx context.Context, qctx *query.Context) query.Result {
	opts := []read.Option{read.OrderBy("created_at", read.Desc)}
	if q.ProjectID != "" {
		opts = append(opts, read.Eq("project_id", q.ProjectID))
	}

	runs, err := read.FindMany[RunCostRow](ctx, opts...)
	if err != nil {
		return query.Fail("read_error", fmt.Sprintf("cost: %v", err))
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

	return query.OK(query.Extras{
		"breakdown": breakdown,
		"totalCost": total,
		"totalRuns": len(runs),
	})
}
