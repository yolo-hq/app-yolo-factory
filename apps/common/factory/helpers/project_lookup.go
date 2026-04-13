package helpers

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/filter"
	"github.com/yolo-hq/yolo/core/read"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// FindProjectByIDOrName tries ID first, then falls back to name filter.
// Caller must have a DB on ctx (write.WithFullDB) — typically wired by
// the action pipeline or job server.
func FindProjectByIDOrName(ctx context.Context, idOrName string) (*entities.Project, error) {
	if p, err := read.FindOne[entities.Project](ctx, idOrName); err == nil && p.ID != "" {
		return &p, nil
	}

	list, err := read.FindMany[entities.Project](ctx, filter.Eq("name", idOrName))
	if err != nil {
		return nil, fmt.Errorf("find project by name: %w", err)
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("project %q not found", idOrName)
	}
	return &list[0], nil
}
