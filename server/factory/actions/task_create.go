package actions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oklog/ulid/v2"
	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/inputs"
)

type CreateTaskAction struct {
	action.TypedInput[inputs.CreateTaskInput]
	TaskWrite entity.WriteRepository[entities.Task]
	TaskRead  entity.ReadRepository[entities.Task]
}


func (a *CreateTaskAction) Execute(ctx context.Context, actx *action.Context) action.Result {
	input, r := a.Input(actx)
	if r != nil {
		return *r
	}

	// Parse dependsOn
	deps := parseDeps(input.DependsOn)

	// Validate all deps exist
	for _, depID := range deps {
		dep, _ := a.TaskRead.FindOne(ctx, entity.FindOneOptions{ID: depID})
		if dep == nil {
			return action.Failure(fmt.Sprintf("dependency task %s not found", depID))
		}
	}

	// Detect cycles
	if hasCycle(ctx, a.TaskRead, deps) {
		return action.Failure("circular dependency detected")
	}

	// Determine initial status
	status := "queued"
	if len(deps) > 0 {
		for _, depID := range deps {
			dep, _ := a.TaskRead.FindOne(ctx, entity.FindOneOptions{ID: depID})
			if dep != nil && dep.Status != "done" {
				status = "blocked"
				break
			}
		}
	}

	// Defaults
	taskType := input.Type
	if taskType == "" {
		taskType = "auto"
	}
	priority := input.Priority
	if priority == 0 {
		priority = 3
	}
	maxRetries := input.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}
	timeoutSecs := input.TimeoutSecs
	if timeoutSecs == 0 {
		timeoutSecs = 600
	}
	labels := input.Labels
	if labels == "" {
		labels = "[]"
	}
	dependsOn := input.DependsOn
	if dependsOn == "" {
		dependsOn = "[]"
	}

	task := &entities.Task{
		RepoID:      input.RepoID,
		Title:       input.Title,
		Body:        input.Body,
		Type:        taskType,
		Status:      status,
		Priority:    priority,
		Model:       input.Model,
		Labels:      labels,
		ParentID:    input.ParentID,
		DependsOn:   dependsOn,
		MaxRetries:  maxRetries,
		TimeoutSecs: timeoutSecs,
	}
	task.ID = ulid.Make().String()

	created, err := a.TaskWrite.Insert(ctx, task)
	if err != nil {
		return action.InternalError()
	}
	return action.Success(created, "task created")
}

func parseDeps(raw string) []string {
	if raw == "" || raw == "[]" {
		return nil
	}
	var deps []string
	json.Unmarshal([]byte(raw), &deps)
	return deps
}

// hasCycle checks if adding deps would create a circular dependency via DFS.
func hasCycle(ctx context.Context, repo entity.ReadRepository[entities.Task], newDeps []string) bool {
	visited := make(map[string]bool)
	newDepSet := make(map[string]bool)
	for _, d := range newDeps {
		newDepSet[d] = true
	}

	var dfs func(id string) bool
	dfs = func(id string) bool {
		if visited[id] {
			return false
		}
		visited[id] = true

		task, _ := repo.FindOne(ctx, entity.FindOneOptions{ID: id})
		if task == nil {
			return false
		}

		for _, d := range parseDeps(task.DependsOn) {
			if newDepSet[d] {
				return true
			}
			if dfs(d) {
				return true
			}
		}
		return false
	}

	for _, dep := range newDeps {
		if dfs(dep) {
			return true
		}
	}
	return false
}

// allDepsDone checks if all dependency tasks have status "done".
func allDepsDone(ctx context.Context, repo entity.ReadRepository[entities.Task], deps []string) bool {
	for _, depID := range deps {
		dep, _ := repo.FindOne(ctx, entity.FindOneOptions{ID: depID})
		if dep == nil || dep.Status != "done" {
			return false
		}
	}
	return true
}
