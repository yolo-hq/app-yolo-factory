package services

import (
	"context"
	"fmt"

	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/read"
	"github.com/yolo-hq/yolo/core/service"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/helpers"
)

// DependencyService validates task dependencies and unblocks tasks when deps complete.
type DependencyService struct {
	service.Base
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
	task, err := read.FindOne[entities.Task](ctx, in.TaskID)
	if err != nil {
		return DependencyOutput{}, fmt.Errorf("load task %s: %w", in.TaskID, err)
	}
	if task.ID == "" {
		return DependencyOutput{Valid: false, CycleError: fmt.Sprintf("task %s not found", in.TaskID)}, nil
	}

	// Load only tasks from the same PRD for cycle detection.
	tasks, err := read.FindMany[entities.Task](ctx,
		read.Eq(fields.Task.PrdID.Name(), task.PrdID),
		read.Limit(1000),
	)
	if err != nil {
		return DependencyOutput{}, fmt.Errorf("load tasks for prd %s: %w", task.PrdID, err)
	}

	taskMap := make(map[string]*entities.Task, len(tasks))
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
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
		if t.Status != string(enums.TaskStatusDone) {
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
	blocked, err := read.FindMany[entities.Task](ctx,
		read.Eq(fields.Task.Status.Name(), string(enums.TaskStatusBlocked)),
		read.Limit(1000),
	)
	if err != nil {
		return out, fmt.Errorf("find blocked tasks: %w", err)
	}

	for _, task := range blocked {
		deps := helpers.ParseDeps(task.DependsOn)
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
			Set(fields.Task.Status.Name(), string(enums.TaskStatusQueued)).
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
		t, err := read.FindOne[entities.Task](ctx, id)
		if err != nil {
			return false, err
		}
		if t.ID == "" || t.Status != string(enums.TaskStatusDone) {
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
	allTasks[taskID] = &entities.Task{DependsOn: helpers.ToJSON(dependsOn)}
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

		for _, depID := range helpers.ParseDeps(t.DependsOn) {
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

func containsStr(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func (s *DependencyService) Description() string {
	return "Validate task dependencies and unblock tasks"
}
