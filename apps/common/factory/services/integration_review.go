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

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/constants"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/helpers"
)

// defaultIntegrationReviewEvery is how often (in completed tasks) to trigger integration review.
const defaultIntegrationReviewEvery = 5

// IntegrationReviewService spawns a read-only agent to review cross-task integration.
type IntegrationReviewService struct {
	service.Base
	Claude  *claude.Client
	Context *ContextService
}

// IntegrationReviewInput holds the data needed for integration review.
type IntegrationReviewInput struct {
	Project      entities.Project
	RecentTasks  []entities.Task
	CombinedDiff string
}

// IntegrationReviewOutput holds the review results.
type IntegrationReviewOutput struct {
	Findings    []IntegrationFinding
	Suggestions []entities.Suggestion
}

// IntegrationFinding represents a single cross-task integration issue.
type IntegrationFinding struct {
	Category string   `json:"category"` // duplicate_function, inconsistent_pattern, state_machine_drift, missing_helper
	Severity string   `json:"severity"` // error, warning, info
	Message  string   `json:"message"`
	Files    []string `json:"files,omitempty"`
}

// integrationReviewOutput wraps the agent's JSON schema output.
type integrationReviewOutput struct {
	Findings []IntegrationFinding `json:"findings"`
}

// Execute runs the integration review agent and returns findings + suggestions.
func (s *IntegrationReviewService) Execute(ctx context.Context, in IntegrationReviewInput) (IntegrationReviewOutput, error) {
	// 1. Build prompt via ContextService.
	ctxOut, err := s.Context.Execute(ctx, ContextInput{
		Phase:         "integration_review",
		Project:       in.Project,
		TaskSummaries: formatTaskSummaries(in.RecentTasks),
		GitDiff:       in.CombinedDiff,
	})
	if err != nil {
		return IntegrationReviewOutput{}, fmt.Errorf("build context: %w", err)
	}

	// 2. Spawn Sonnet agent (read-only, bare, budget $0.50, plan mode).
	result, err := s.Claude.Run(ctx, claude.Config{
		Model:          "sonnet",
		AllowedTools:   []string{"Read", "Glob", "Grep"},
		Bare:           true,
		BudgetUSD:      0.50,
		PermissionMode: "auto",
		Effort:         "medium",
		CWD:            in.Project.LocalPath,
		JSONSchema:     constants.IntegrationReviewSchema,
		SessionName:    fmt.Sprintf("factory:project-%s:integration-review", in.Project.ID),
		Timeout:        5 * time.Minute,
	}, ctxOut.Prompt)
	if err != nil {
		return IntegrationReviewOutput{}, fmt.Errorf("claude run: %w", err)
	}

	if result.IsError {
		return IntegrationReviewOutput{}, fmt.Errorf("claude error: %s", result.Text)
	}

	// 3. Parse structured output into findings.
	findings, err := parseIntegrationReviewOutput(result.StructuredOutput)
	if err != nil {
		return IntegrationReviewOutput{}, fmt.Errorf("parse output: %w", err)
	}

	// 4. Convert high-priority findings to Suggestion entities.
	suggestions := findingsToSuggestions(findings, in.Project.ID)

	return IntegrationReviewOutput{
		Findings:    findings,
		Suggestions: suggestions,
	}, nil
}

// ShouldRunIntegrationReview returns true if it's time for an integration review.
// Triggers every N completed tasks (default: 5).
func ShouldRunIntegrationReview(completedCount int, every int) bool {
	if every <= 0 {
		every = defaultIntegrationReviewEvery
	}
	return completedCount > 0 && completedCount%every == 0
}

// parseIntegrationReviewOutput parses the agent's structured JSON output.
func parseIntegrationReviewOutput(raw json.RawMessage) ([]IntegrationFinding, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty structured output")
	}

	var out integrationReviewOutput
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("unmarshal findings: %w", err)
	}
	return out.Findings, nil
}

// findingsToSuggestions converts error/warning findings into Suggestion entities.
func findingsToSuggestions(findings []IntegrationFinding, projectID string) []entities.Suggestion {
	var suggestions []entities.Suggestion
	for _, f := range findings {
		if f.Severity == "info" {
			continue
		}
		priority := string(enums.SuggestionPriorityMedium)
		if f.Severity == "error" {
			priority = string(enums.SuggestionPriorityHigh)
		}
		s := entities.Suggestion{
			ProjectID: projectID,
			Source:    "integration_review",
			Category:  mapFindingCategory(f.Category),
			Title:     fmt.Sprintf("[%s] %s", f.Category, helpers.Truncate(f.Message, 80)),
			Body:      f.Message,
			Priority:  priority,
			Status:    string(enums.SuggestionStatusPending),
		}
		s.ID = ulid.Make().String()
		suggestions = append(suggestions, s)
	}
	return suggestions
}

// mapFindingCategory maps integration finding categories to suggestion categories.
func mapFindingCategory(category string) string {
	switch category {
	case "duplicate_function", "missing_helper":
		return string(enums.SuggestionCategoryRefactoring)
	case "inconsistent_pattern", "state_machine_drift":
		return string(enums.SuggestionCategoryTechDebt)
	default:
		return string(enums.SuggestionCategoryRefactoring)
	}
}

// formatTaskSummaries formats completed tasks for template injection.
func formatTaskSummaries(tasks []entities.Task) string {
	if len(tasks) == 0 {
		return "No recent tasks."
	}

	var lines []string
	for _, t := range tasks {
		line := fmt.Sprintf("- [%s] %s: %s", t.ID, t.Title, helpers.Truncate(t.Spec, 200))
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (s *IntegrationReviewService) Description() string {
	return "Review cross-task integration patterns"
}
