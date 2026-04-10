package commands

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// findProjectByIDOrName tries ID first, then falls back to name filter.
func findProjectByIDOrName(ctx context.Context, r entity.ReadRepository[entities.Project], idOrName string) (*entities.Project, error) {
	p, err := r.FindOne(ctx, entity.FindOneOptions{ID: idOrName})
	if err != nil {
		return nil, fmt.Errorf("find project: %w", err)
	}
	if p != nil {
		return p, nil
	}

	p, err = r.FindOne(ctx, entity.FindOneOptions{
		Filters: []entity.FilterCondition{
			{Field: "name", Operator: entity.OpEq, Value: idOrName},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("find project by name: %w", err)
	}
	if p == nil {
		return nil, fmt.Errorf("project %q not found", idOrName)
	}
	return p, nil
}
