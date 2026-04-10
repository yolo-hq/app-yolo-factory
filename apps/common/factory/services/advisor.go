package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/pkg/claude"
	"github.com/yolo-hq/yolo/core/service"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/constants"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/helpers"
)

// AdvisorService spawns a read-only agent to analyze a project and suggest improvements.
type AdvisorService struct {
	service.Base
	Claude          *claude.Client
	Context         *ContextService
	ProjectRead     entity.ReadRepository[entities.Project]
	TaskRead        entity.ReadRepository[entities.Task]
	RunRead         entity.ReadRepository[entities.Run]
	SuggestionWrite entity.WriteRepository[entities.Suggestion]
}

// AdvisorInput holds the data needed for advisor analysis.
type AdvisorInput struct {
	ProjectID    string
	AnalysisType string // "pattern_extraction", "code_quality", "performance", "architecture", "model_optimization"
}

// AdvisorOutput holds the generated suggestions.
type AdvisorOutput struct {
	Suggestions []entities.Suggestion
}

// advisorSuggestionDef is the structured output from the advisor agent.
type advisorSuggestionDef struct {
	Title           string `json:"title"`
	Body            string `json:"body"`
	Category        string `json:"category"`
	Priority        string `json:"priority"`
	EstimatedImpact string `json:"estimated_impact"`
}

// advisorOutput wraps the agent's JSON schema output.
type advisorOutput struct {
	Suggestions []advisorSuggestionDef `json:"suggestions"`
}

// Execute loads the project, runs the advisor agent, persists suggestions, and returns them.
func (s *AdvisorService) Execute(ctx context.Context, in AdvisorInput) (AdvisorOutput, error) {
	// Load project.
	project, err := s.ProjectRead.FindOne(ctx, entity.FindOneOptions{ID: in.ProjectID})
	if err != nil {
		return AdvisorOutput{}, fmt.Errorf("load project: %w", err)
	}
	if project == nil {
		return AdvisorOutput{}, fmt.Errorf("project %s not found", in.ProjectID)
	}

	// Load tasks for this project.
	taskResult, err := s.TaskRead.FindMany(ctx, entity.FindOptions{
		Filters: []entity.FilterCondition{
			{Field: "project_id", Operator: entity.OpEq, Value: in.ProjectID},
		},
	})
	if err != nil {
		return AdvisorOutput{}, fmt.Errorf("load tasks: %w", err)
	}

	// Collect task IDs to scope runs to this project.
	taskIDs := make([]any, len(taskResult.Data))
	for i, t := range taskResult.Data {
		taskIDs[i] = t.ID
	}
	if len(taskIDs) == 0 {
		// No tasks means no runs to analyze.
		return AdvisorOutput{}, nil
	}

	// Load recent runs scoped to this project's tasks.
	runResult, err := s.RunRead.FindMany(ctx, entity.FindOptions{
		Filters: []entity.FilterCondition{
			{Field: "task_id", Operator: entity.OpIn, Value: taskIDs},
		},
		Pagination: &entity.PaginationParams{Limit: 20},
		Sort:       &entity.SortParams{Field: "created_at", Order: "desc"},
	})
	if err != nil {
		return AdvisorOutput{}, fmt.Errorf("load runs: %w", err)
	}

	// 1. Build prompt via ContextService.
	claudeMD := readCLAUDEMD(project.LocalPath)

	ctxOut, err := s.Context.Execute(ctx, ContextInput{
		Phase:           "advisor",
		Project:         *project,
		CLAUDEMDContent: claudeMD,
		AnalysisType:    in.AnalysisType,
		AnalysisContext: claudeMD,
		RunHistory:      formatRunHistory(runResult.Data),
	})
	if err != nil {
		return AdvisorOutput{}, fmt.Errorf("build context: %w", err)
	}

	// 2. Spawn Sonnet agent (read-only, NOT bare — needs project memory).
	result, err := s.Claude.Run(ctx, claude.Config{
		Model:          "sonnet",
		AllowedTools:   []string{"Read", "Glob", "Grep"},
		Bare:           false,
		BudgetUSD:      0.50,
		PermissionMode: "auto",
		Effort:         "medium",
		CWD:            project.LocalPath,
		JSONSchema:     constants.AdvisorSchema,
		SessionName:    fmt.Sprintf("factory:project-%s:advisor", project.ID),
		Timeout:        10 * time.Minute,
	}, ctxOut.Prompt)
	if err != nil {
		return AdvisorOutput{}, fmt.Errorf("claude run: %w", err)
	}

	if result.IsError {
		return AdvisorOutput{}, fmt.Errorf("claude error: %s", result.Text)
	}

	// 3. Parse structured output.
	defs, err := parseAdvisorOutput(result.StructuredOutput)
	if err != nil {
		return AdvisorOutput{}, fmt.Errorf("parse output: %w", err)
	}

	// 4. Convert to Suggestion entity structs and persist.
	suggestions := make([]entities.Suggestion, len(defs))
	for i, def := range defs {
		suggestions[i] = entities.Suggestion{
			ProjectID: project.ID,
			Source:    "advisor",
			Category:  mapAdvisorCategory(def.Category),
			Title:     def.Title,
			Body:      def.Body,
			Priority:  def.Priority,
		}
		suggestions[i].ID = ulid.Make().String()
		if _, err := s.SuggestionWrite.Insert(ctx, &suggestions[i]); err != nil {
			return AdvisorOutput{}, fmt.Errorf("insert suggestion: %w", err)
		}
	}

	return AdvisorOutput{Suggestions: suggestions}, nil
}

// parseAdvisorOutput parses the agent's structured JSON output.
func parseAdvisorOutput(raw json.RawMessage) ([]advisorSuggestionDef, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty structured output")
	}

	var out advisorOutput
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("unmarshal suggestions: %w", err)
	}
	return out.Suggestions, nil
}

// mapAdvisorCategory maps advisor schema categories to entity categories.
func mapAdvisorCategory(category string) string {
	switch category {
	case "optimization", "refactoring", "tech_debt":
		return "refactor"
	case "new_feature", "pattern_extraction":
		return "feature"
	default:
		return "refactor"
	}
}

// formatRunHistory formats run entities for template injection.
func formatRunHistory(runs []entities.Run) string {
	if len(runs) == 0 {
		return "No recent runs."
	}

	var lines []string
	for _, r := range runs {
		line := fmt.Sprintf("- Run %s: status=%s model=%s cost=$%.2f", r.ID, r.Status, r.Model, r.CostUSD)
		if r.Error != "" {
			line += fmt.Sprintf(" error=%s", helpers.Truncate(r.Error, 100))
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (s *AdvisorService) Description() string {
	return "Run optimization advisor analysis on project metrics"
}
