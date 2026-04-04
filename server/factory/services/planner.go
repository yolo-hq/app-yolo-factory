package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/yolo-hq/yolo/core/pkg/claude"
	"github.com/yolo-hq/yolo/core/service"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/server/factory/skills"
)

// PlannerService spawns a Claude agent to break a PRD into implementation tasks.
type PlannerService struct {
	service.Base
	Claude     *claude.Client
	Context    *ContextService
	Dependency *DependencyService
}

// PlannerInput holds the data needed for planning.
type PlannerInput struct {
	PRD     entities.PRD
	Project entities.Project
}

// PlannerOutput holds the planned tasks (not persisted — caller persists).
type PlannerOutput struct {
	Tasks []entities.Task
	Count int
}

// TaskDef is the structured output from the planner agent.
type TaskDef struct {
	Title              string        `json:"title"`
	Spec               string        `json:"spec"`
	AcceptanceCriteria []CriteriaDef `json:"acceptance_criteria"`
	Branch             string        `json:"branch"`
	Sequence           int           `json:"sequence"`
	DependsOn          []int         `json:"depends_on"`
	EstimatedComplexity string       `json:"estimated_complexity"`
}

// CriteriaDef is a single acceptance criterion from the agent.
type CriteriaDef struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// planTasksOutput wraps the agent's JSON schema output.
type planTasksOutput struct {
	Tasks []TaskDef `json:"tasks"`
}

// Execute runs the planner agent and returns task entities.
func (s *PlannerService) Execute(ctx context.Context, in PlannerInput) (PlannerOutput, error) {
	// 1. Build context prompt.
	claudeMD := readCLAUDEMD(in.Project.LocalPath)

	ctxOut, err := s.Context.Execute(ctx, ContextInput{
		Phase:           "plan_tasks",
		PRD:             in.PRD,
		Project:         in.Project,
		CLAUDEMDContent: claudeMD,
	})
	if err != nil {
		return PlannerOutput{}, fmt.Errorf("build context: %w", err)
	}

	// 2. Spawn planner agent.
	result, err := s.Claude.Run(ctx, claude.Config{
		Model:          "opus",
		AllowedTools:   []string{"Read", "Glob", "Grep"},
		Bare:           true,
		BudgetUSD:      1.0,
		PermissionMode: "auto",
		Effort:         "high",
		CWD:            in.Project.LocalPath,
		JSONSchema:     skills.PlanTasksSchema,
		SessionName:    fmt.Sprintf("factory:prd-%s:plan", in.PRD.ID),
		Timeout:        10 * time.Minute,
	}, ctxOut.Prompt)
	if err != nil {
		return PlannerOutput{}, fmt.Errorf("claude run: %w", err)
	}

	if result.IsError {
		return PlannerOutput{}, fmt.Errorf("claude error: %s", result.Text)
	}

	// 3. Parse structured output.
	defs, err := parseTaskDefs(result.StructuredOutput)
	if err != nil {
		return PlannerOutput{}, fmt.Errorf("parse output: %w", err)
	}

	// 4. Convert to entities.
	tasks, err := s.convertToEntities(defs, in)
	if err != nil {
		return PlannerOutput{}, fmt.Errorf("convert tasks: %w", err)
	}

	// 5. Validate dependencies (cycle detection).
	for _, task := range tasks {
		deps := parseDepsJSON(task.DependsOn)
		if len(deps) == 0 {
			continue
		}
		depOut, err := s.Dependency.Execute(ctx, DependencyInput{
			TaskID:    task.ID,
			DependsOn: deps,
		})
		if err != nil {
			return PlannerOutput{}, fmt.Errorf("validate deps for task %s: %w", task.Title, err)
		}
		if depOut.CycleError != "" {
			return PlannerOutput{}, fmt.Errorf("cycle in task %s: %s", task.Title, depOut.CycleError)
		}
	}

	return PlannerOutput{
		Tasks: tasks,
		Count: len(tasks),
	}, nil
}

// parseTaskDefs parses the agent's structured JSON output.
func parseTaskDefs(raw json.RawMessage) ([]TaskDef, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty structured output")
	}

	var out planTasksOutput
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("unmarshal tasks: %w", err)
	}
	if len(out.Tasks) == 0 {
		return nil, fmt.Errorf("no tasks in output")
	}
	return out.Tasks, nil
}

// convertToEntities turns TaskDefs into entity.Task structs with IDs and deps mapped.
func (s *PlannerService) convertToEntities(defs []TaskDef, in PlannerInput) ([]entities.Task, error) {
	// Generate IDs and build sequence→ID map.
	seqToID := make(map[int]string, len(defs))
	tasks := make([]entities.Task, len(defs))

	for i, def := range defs {
		id := ulid.Make().String()
		seqToID[def.Sequence] = id

		acJSON, err := json.Marshal(def.AcceptanceCriteria)
		if err != nil {
			return nil, fmt.Errorf("marshal criteria for task %d: %w", def.Sequence, err)
		}

		branch := def.Branch
		if branch == "" {
			branch = in.Project.DefaultBranch
		}

		tasks[i] = entities.Task{
			Title:              def.Title,
			Spec:               def.Spec,
			AcceptanceCriteria: string(acJSON),
			Branch:             branch,
			Sequence:           def.Sequence,
			PrdID:              in.PRD.ID,
			ProjectID:          in.Project.ID,
		}
		tasks[i].ID = id
	}

	// Map depends_on from sequence numbers to task IDs.
	for i, def := range defs {
		if len(def.DependsOn) == 0 {
			tasks[i].DependsOn = "[]"
			tasks[i].Status = "queued"
			continue
		}

		depIDs := make([]string, 0, len(def.DependsOn))
		for _, seq := range def.DependsOn {
			depID, ok := seqToID[seq]
			if !ok {
				return nil, fmt.Errorf("task %d depends on unknown sequence %d", def.Sequence, seq)
			}
			depIDs = append(depIDs, depID)
		}

		tasks[i].DependsOn = toJSON(depIDs)
		tasks[i].Status = "blocked"
	}

	return tasks, nil
}

// readCLAUDEMD reads CLAUDE.md from a project path. Returns empty string on error.
func readCLAUDEMD(projectPath string) string {
	if projectPath == "" {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(projectPath, "CLAUDE.md"))
	if err != nil {
		return ""
	}
	return string(data)
}
