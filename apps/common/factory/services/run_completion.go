package services

import (
	"context"

	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/service"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/helpers"
)

// RunCompletionService handles graph traversal for run completion:
// cascading failure to dependents and checking whether a task's
// dependencies are all satisfied.
type RunCompletionService struct {
	service.Base
	TaskRead  entity.ReadRepository[entities.Task]
	TaskWrite entity.WriteRepository[entities.Task]
}

func (s *RunCompletionService) Description() string {
	return "Cascade failures and check task dependencies for run completion"
}

// CascadeFailure recursively marks non-terminal tasks that depend on
// failedTaskID (directly or transitively) as failed, scoped to prdID.
func (s *RunCompletionService) CascadeFailure(ctx context.Context, failedTaskID, prdID string) error {
	// Load all tasks in the same PRD and walk dependents.
	result, err := s.TaskRead.FindMany(ctx, entity.FindOptions{
		Filters: []entity.FilterCondition{
			{Field: "prd_id", Operator: entity.OpEq, Value: prdID},
		},
	})
	if err != nil {
		return err
	}

	for _, t := range result.Data {
		if t.Status == string(enums.TaskStatusDone) || t.Status == string(enums.TaskStatusFailed) || t.Status == string(enums.TaskStatusCancelled) {
			continue
		}
		if !helpers.ContainsDep(t.DependsOn, failedTaskID) {
			continue
		}
		if _, err := s.TaskWrite.Update(ctx).
			WhereID(t.ID).
			Set(fields.Task.Status.Name(), string(enums.TaskStatusFailed)).
			Exec(ctx); err != nil {
			continue
		}
		// Recurse: cascade to tasks depending on this one.
		if err := s.CascadeFailure(ctx, t.ID, prdID); err != nil {
			return err
		}
	}
	return nil
}

// AllDepsMet returns true when every dep ID has status "done".
func (s *RunCompletionService) AllDepsMet(ctx context.Context, depIDs []string) (bool, error) {
	for _, id := range depIDs {
		t, err := s.TaskRead.FindOne(ctx, entity.FindOneOptions{ID: id})
		if err != nil {
			return false, err
		}
		if t == nil || t.Status != string(enums.TaskStatusDone) {
			return false, nil
		}
	}
	return true, nil
}
