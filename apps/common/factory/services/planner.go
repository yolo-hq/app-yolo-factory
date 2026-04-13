package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/pkg/claude"
	"github.com/yolo-hq/yolo/core/read"
	"github.com/yolo-hq/yolo/core/service"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/constants"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/helpers"
)

// PlannerService spawns a Claude agent to break a PRD into implementation tasks.
type PlannerService struct {
	service.Base
	Claude     *claude.Client
	Context    *ContextService
	Dependency *DependencyService
	PRDWrite   entity.WriteRepository[entities.PRD]
	TaskWrite  entity.WriteRepository[entities.Task]
}

// PlannerInput holds the data needed for planning.
type PlannerInput struct {
	PRDID string
}

// PlannerOutput holds the planned tasks (not persisted — caller persists).
type PlannerOutput struct {
	Tasks []entities.Task
	Count int
}

// TaskDef is the structured output from the planner agent.
type TaskDef struct {
	Title               string        `json:"title"`
	Spec                string        `json:"spec"`
	AcceptanceCriteria  []CriteriaDef `json:"acceptance_criteria"`
	Branch              string        `json:"branch"`
	Sequence            int           `json:"sequence"`
	DependsOn           []int         `json:"depends_on"`
	EstimatedComplexity string        `json:"estimated_complexity"`
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

// Execute loads the PRD and project, runs the planner agent, persists tasks, and updates the PRD.
func (s *PlannerService) Execute(ctx context.Context, in PlannerInput) (PlannerOutput, error) {
	// 1. Load PRD.
	prd, err := read.FindOne[entities.PRD](ctx, in.PRDID)
	if err != nil {
		return PlannerOutput{}, fmt.Errorf("load prd: %w", err)
	}
	if prd.ID == "" {
		return PlannerOutput{}, fmt.Errorf("prd %s not found", in.PRDID)
	}

	// 2. Load Project.
	project, err := read.FindOne[entities.Project](ctx, prd.ProjectID)
	if err != nil {
		return PlannerOutput{}, fmt.Errorf("load project: %w", err)
	}
	if project.ID == "" {
		return PlannerOutput{}, fmt.Errorf("project %s not found", prd.ProjectID)
	}

	// 3. Build context prompt.
	claudeMD := readCLAUDEMD(project.LocalPath)

	ctxOut, err := s.Context.Execute(ctx, ContextInput{
		Phase:           "plan_tasks",
		PRD:             prd,
		Project:         project,
		CLAUDEMDContent: claudeMD,
	})
	if err != nil {
		return PlannerOutput{}, fmt.Errorf("build context: %w", err)
	}

	// 4. Spawn planner agent.
	result, err := s.Claude.Run(ctx, claude.Config{
		Model:          "opus",
		AllowedTools:   []string{"Read", "Glob", "Grep"},
		Bare:           true,
		BudgetUSD:      1.0,
		PermissionMode: "auto",
		Effort:         "high",
		CWD:            project.LocalPath,
		JSONSchema:     constants.PlanTasksSchema,
		SessionName:    fmt.Sprintf("factory:prd-%s:plan", prd.ID),
		Timeout:        10 * time.Minute,
	}, ctxOut.Prompt)
	if err != nil {
		s.markPRDFailed(ctx, prd.ID)
		return PlannerOutput{}, fmt.Errorf("claude run: %w", err)
	}

	if result.IsError {
		s.markPRDFailed(ctx, prd.ID)
		return PlannerOutput{}, fmt.Errorf("claude error: %s", result.Text)
	}

	// 5. Parse structured output.
	defs, err := parseTaskDefs(result.StructuredOutput)
	if err != nil {
		s.markPRDFailed(ctx, prd.ID)
		return PlannerOutput{}, fmt.Errorf("parse output: %w", err)
	}

	// 6. Convert to entities.
	planInput := plannerEntitiesInput{PRD: prd, Project: project}
	tasks, err := s.convertToEntities(defs, planInput)
	if err != nil {
		s.markPRDFailed(ctx, prd.ID)
		return PlannerOutput{}, fmt.Errorf("convert tasks: %w", err)
	}

	// 7. Validate dependencies (cycle detection).
	for _, task := range tasks {
		deps := helpers.ParseDeps(task.DependsOn)
		if len(deps) == 0 {
			continue
		}
		depOut, err := s.Dependency.Execute(ctx, DependencyInput{
			TaskID:    task.ID,
			DependsOn: deps,
		})
		if err != nil {
			s.markPRDFailed(ctx, prd.ID)
			return PlannerOutput{}, fmt.Errorf("validate deps for task %s: %w", task.Title, err)
		}
		if depOut.CycleError != "" {
			s.markPRDFailed(ctx, prd.ID)
			return PlannerOutput{}, fmt.Errorf("cycle in task %s: %s", task.Title, depOut.CycleError)
		}
	}

	// 8. Persist tasks.
	for i := range tasks {
		if _, err := s.TaskWrite.Insert(ctx, &tasks[i]); err != nil {
			return PlannerOutput{}, fmt.Errorf("insert task %s: %w", tasks[i].Title, err)
		}
	}

	// 9. Update PRD: set total_tasks and transition to approved.
	_, err = s.PRDWrite.Update(ctx).
		WhereID(prd.ID).
		Set(fields.PRD.Status.Name(), string(enums.PRDStatusApproved)).
		Exec(ctx)
	if err != nil {
		return PlannerOutput{}, fmt.Errorf("update prd: %w", err)
	}

	return PlannerOutput{
		Tasks: tasks,
		Count: len(tasks),
	}, nil
}

// markPRDFailed marks a PRD as failed (best-effort).
func (s *PlannerService) markPRDFailed(ctx context.Context, prdID string) {
	if _, err := s.PRDWrite.Update(ctx).
		WhereID(prdID).
		Set(fields.PRD.Status.Name(), string(enums.PRDStatusFailed)).
		Exec(ctx); err != nil {
		slog.Error("failed to mark PRD as failed", "prd_id", prdID, "error", err)
	}
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

// plannerEntitiesInput holds loaded entities for task conversion.
type plannerEntitiesInput struct {
	PRD     entities.PRD
	Project entities.Project
}

// convertToEntities turns TaskDefs into entity.Task structs with IDs and deps mapped.
func (s *PlannerService) convertToEntities(defs []TaskDef, in plannerEntitiesInput) ([]entities.Task, error) {
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
			tasks[i].Status = string(enums.TaskStatusQueued)
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

		tasks[i].DependsOn = helpers.ToJSON(depIDs)
		tasks[i].Status = string(enums.TaskStatusBlocked)
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

func (s *PlannerService) Description() string { return "Break a PRD into implementable tasks" }
