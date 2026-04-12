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

// InsightRow is the projection used for list responses.
type InsightRow struct {
	projection.For[entities.Insight]

	ID       string `field:"id"`
	Title    string `field:"title"`
	Category string `field:"category"`
	Priority string `field:"priority"`
	Status   string `field:"status"`
}

// ListInsightsResponse is the typed response for ListInsightsAction.
type ListInsightsResponse struct {
	Insights []InsightRow `json:"insights"`
}

// ListInsightsAction lists insights with optional filters.
type ListInsightsAction struct {
	action.SkipAllPolicies
	action.TypedInput[inputs.ListInsightsInput]
	action.TypedResponse[ListInsightsResponse]
}

func (a *ListInsightsAction) Description() string { return "List insights with optional filters" }

func (a *ListInsightsAction) Execute(ctx context.Context, actx *action.Context) error {
	input := a.Input(actx)

	opts := []read.Option{read.OrderBy("created_at", read.Desc)}
	if input.ProjectID != "" {
		opts = append(opts, read.Eq("project_id", input.ProjectID))
	}
	if input.Category != "" {
		opts = append(opts, read.Eq("category", input.Category))
	}
	if input.Priority != "" {
		opts = append(opts, read.Eq("priority", input.Priority))
	}
	if input.Status != "" {
		opts = append(opts, read.Eq("status", input.Status))
	}

	insights, err := read.FindMany[InsightRow](ctx, opts...)
	if err != nil {
		return fmt.Errorf("list-insights: %w", err)
	}

	return a.Respond(actx, ListInsightsResponse{Insights: insights})
}
