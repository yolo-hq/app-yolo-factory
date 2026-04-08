package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/yolo-hq/yolo/core/pkg/claude"
	"github.com/yolo-hq/yolo/core/service"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/constants"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/helpers"
)

// AdvisorService spawns a read-only agent to analyze a project and suggest improvements.
type AdvisorService struct {
	service.Base
	Claude  *claude.Client
	Context *ContextService
}

// AdvisorInput holds the data needed for advisor analysis.
type AdvisorInput struct {
	Project      entities.Project
	AnalysisType string // "pattern_extraction", "code_quality", "performance", "architecture", "model_optimization"
	RunHistory   []entities.Run
}

// AdvisorOutput holds the generated suggestions (not persisted — caller persists).
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

// Execute runs the advisor agent and returns suggestion entities.
func (s *AdvisorService) Execute(ctx context.Context, in AdvisorInput) (AdvisorOutput, error) {
	// 1. Build prompt via ContextService.
	claudeMD := readCLAUDEMD(in.Project.LocalPath)

	ctxOut, err := s.Context.Execute(ctx, ContextInput{
		Phase:           "advisor",
		Project:         in.Project,
		CLAUDEMDContent: claudeMD,
		AnalysisType:    in.AnalysisType,
		AnalysisContext: claudeMD,
		RunHistory:      formatRunHistory(in.RunHistory),
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
		CWD:            in.Project.LocalPath,
		JSONSchema:     constants.AdvisorSchema,
		SessionName:    fmt.Sprintf("factory:project-%s:advisor", in.Project.ID),
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

	// 4. Convert to Suggestion entity structs.
	suggestions := make([]entities.Suggestion, len(defs))
	for i, def := range defs {
		suggestions[i] = entities.Suggestion{
			ProjectID: in.Project.ID,
			Source:    "advisor",
			Category:  mapAdvisorCategory(def.Category),
			Title:     def.Title,
			Body:      def.Body,
			Priority:  def.Priority,
		}
		suggestions[i].ID = ulid.Make().String()
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

func (s *AdvisorService) Description() string { return "Run optimization advisor analysis on project metrics" }
