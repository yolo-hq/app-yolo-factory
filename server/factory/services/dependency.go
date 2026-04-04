package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/service"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

// DependencyService validates task dependencies and unblocks tasks when deps complete.
type DependencyService struct {
	service.Base
	TaskRead  entity.ReadRepository[entities.Task]
	TaskWrite entity.WriteRepository[entities.Task]
}

// DependencyInput is the input for dependency validation.
type DependencyInput struct {
	TaskID    string
	DependsOn []string
}

// DependencyOutput is the result of dependency validation.
type DependencyOutput struct {
	Valid      bool
	CycleError string
	BlockedBy  []string
	AllMet     bool
}

// Execute validates that all dependencies exist, detects cycles, and checks completion status.
func (s *DependencyService) Execute(ctx context.Context, in DependencyInput) (DependencyOutput, error) {
	out := DependencyOutput{Valid: true, AllMet: true}

	if len(in.DependsOn) == 0 {
		return out, nil
	}

	// Self-dependency check.
	for _, dep := range in.DependsOn {
		if dep == in.TaskID {
			return DependencyOutput{CycleError: fmt.Sprintf("task %s depends on itself", in.TaskID)}, nil
		}
	}

	// Load the task being checked to get its PRD scope.
	task, err := s.TaskRead.FindOne(ctx, entity.FindOneOptions{ID: in.TaskID})
	if err != nil {
		return DependencyOutput{}, fmt.Errorf("load task %s: %w", in.TaskID, err)
	}
	if task == nil {
		return DependencyOutput{Valid: false, CycleError: fmt.Sprintf("task %s not found", in.TaskID)}, nil
	}

	// Load only tasks from the same PRD for cycle detection.
	result, err := s.TaskRead.FindMany(ctx, entity.FindOptions{
		Filters: []entity.FilterCondition{
			{Field: "prd_id", Operator: entity.OpEq, Value: task.PrdID},
		},
	})
	if err != nil {
		return DependencyOutput{}, fmt.Errorf("load tasks for prd %s: %w", task.PrdID, err)
	}

	taskMap := make(map[string]*entities.Task, len(result.Data))
	for i := range result.Data {
		taskMap[result.Data[i].ID] = &result.Data[i]
	}

	// Validate all deps exist.
	for _, depID := range in.DependsOn {
		if _, ok := taskMap[depID]; !ok {
			return DependencyOutput{Valid: false, CycleError: fmt.Sprintf("dependency %s does not exist", depID)}, nil
		}
	}

	// Cycle detection via DFS.
	if err := detectCycle(in.TaskID, in.DependsOn, taskMap); err != nil {
		return DependencyOutput{CycleError: err.Error()}, nil
	}

	// Check which deps are not done.
	for _, depID := range in.DependsOn {
		t := taskMap[depID]
		if t.Status != "done" {
			out.AllMet = false
			out.BlockedBy = append(out.BlockedBy, depID)
		}
	}

	return out, nil
}

// UnblockInput is the input for the Unblock method.
type UnblockInput struct {
	CompletedTaskID string
}

// UnblockOutput is the result of unblocking tasks.
type UnblockOutput struct {
	UnblockedTaskIDs []string
}

// Unblock finds blocked tasks that depend on the completed task and transitions
// them to "queued" when all their dependencies are met.
func (s *DependencyService) Unblock(ctx context.Context, in UnblockInput) (UnblockOutput, error) {
	var out UnblockOutput

	// Load all blocked tasks.
	result, err := s.TaskRead.FindMany(ctx, entity.FindOptions{
		Filters: []entity.FilterCondition{
			{Field: "status", Operator: entity.OpEq, Value: "blocked"},
		},
	})
	if err != nil {
		return out, fmt.Errorf("find blocked tasks: %w", err)
	}

	for _, task := range result.Data {
		deps := parseDepsJSON(task.DependsOn)
		if !containsStr(deps, in.CompletedTaskID) {
			continue
		}

		// Check if ALL deps are done.
		allDone, err := s.allDepsDone(ctx, deps)
		if err != nil {
			return out, fmt.Errorf("check deps for task %s: %w", task.ID, err)
		}
		if !allDone {
			continue
		}

		// Transition blocked → queued.
		_, err = s.TaskWrite.Update(ctx).
			WhereID(task.ID).
			Set("status", "queued").
			Exec(ctx)
		if err != nil {
			return out, fmt.Errorf("unblock task %s: %w", task.ID, err)
		}
		out.UnblockedTaskIDs = append(out.UnblockedTaskIDs, task.ID)
	}

	return out, nil
}

func (s *DependencyService) allDepsDone(ctx context.Context, depIDs []string) (bool, error) {
	for _, id := range depIDs {
		t, err := s.TaskRead.FindOne(ctx, entity.FindOneOptions{ID: id})
		if err != nil {
			return false, err
		}
		if t == nil || t.Status != "done" {
			return false, nil
		}
	}
	return true, nil
}

// detectCycle runs DFS from taskID looking for back-edges.
func detectCycle(taskID string, dependsOn []string, allTasks map[string]*entities.Task) error {
	visited := make(map[string]bool)
	path := make(map[string]bool)

	// Temporarily add the candidate task.
	orig, existed := allTasks[taskID]
	allTasks[taskID] = &entities.Task{DependsOn: toJSON(dependsOn)}
	defer func() {
		if existed {
			allTasks[taskID] = orig
		} else {
			delete(allTasks, taskID)
		}
	}()

	var dfs func(id string) error
	dfs = func(id string) error {
		visited[id] = true
		path[id] = true

		t, ok := allTasks[id]
		if !ok {
			path[id] = false
			return nil
		}

		for _, depID := range parseDepsJSON(t.DependsOn) {
			if path[depID] {
				return fmt.Errorf("cycle detected: %s -> %s", id, depID)
			}
			if !visited[depID] {
				if err := dfs(depID); err != nil {
					return err
				}
			}
		}

		path[id] = false
		return nil
	}

	return dfs(taskID)
}

// parseDepsJSON parses a JSON array string into a slice of strings.
// Returns nil for empty/invalid input.
func parseDepsJSON(jsonStr string) []string {
	if jsonStr == "" || jsonStr == "[]" {
		return nil
	}
	var deps []string
	if err := json.Unmarshal([]byte(jsonStr), &deps); err != nil {
		return nil
	}
	return deps
}

func toJSON(ss []string) string {
	b, _ := json.Marshal(ss)
	return string(b)
}

func containsStr(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
