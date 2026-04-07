package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// Cost shows cost breakdown by project and model.
type Cost struct {
	command.Base
}

type CostInput struct {
	Period  string `flag:"period" usage:"Time period: week or month (default: month)"`
	Project string `flag:"project" usage:"Filter by project ID"`
}

func (c *Cost) Name() string        { return "cost" }
func (c *Cost) Description() string { return "Show cost breakdown" }
func (c *Cost) Input() any          { return &CostInput{} }

func (c *Cost) Execute(ctx context.Context, cctx command.Context) error {
	repo, err := cctx.RepoProvider.Repo("Run")
	if err != nil {
		return fmt.Errorf("get run repo: %w", err)
	}
	r := repo.(entity.ReadRepository[entities.Run])

	result, err := r.FindMany(ctx, entity.FindOptions{
		Sort: &entity.SortParams{Field: "created_at", Order: "desc"},
	})
	if err != nil {
		return fmt.Errorf("list runs: %w", err)
	}

	// Group by model.
	byModel := map[string]float64{}
	var total float64
	for _, run := range result.Data {
		byModel[run.Model] += run.CostUSD
		total += run.CostUSD
	}

	cctx.Print("=== Cost Summary ===")
	cctx.Print("")

	tw := cctx.Table()
	fmt.Fprintf(tw, "MODEL\tCOST\tRUNS\n")

	modelRuns := map[string]int{}
	for _, run := range result.Data {
		modelRuns[run.Model]++
	}
	for model, cost := range byModel {
		fmt.Fprintf(tw, "%s\t$%.2f\t%d\n", model, cost, modelRuns[model])
	}
	fmt.Fprintf(tw, "TOTAL\t$%.2f\t%d\n", total, len(result.Data))
	tw.Flush()

	return nil
}
