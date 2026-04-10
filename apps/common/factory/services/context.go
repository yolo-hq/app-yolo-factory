package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/yolo-hq/yolo/core/service"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/constants"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// ContextService builds prompts for headless agent constants.
// Stateless — caller loads entities and passes them in.
type ContextService struct {
	service.Base
}

// DepSummary is a completed dependency's summary for context injection.
type DepSummary struct {
	Title        string
	Summary      string
	CommitHash   string
	FilesChanged []string
}

// ContextInput holds all data needed to build a skill prompt.
type ContextInput struct {
	Task            entities.Task
	PRD             entities.PRD
	Project         entities.Project
	Phase           string // plan_tasks, implement, review_task, review_prd, audit, sentinel, advisor
	CompletedDeps   []DepSummary
	IsRetry         bool
	RetryError      string
	ReviewReasons   string
	GitDiff         string
	CLAUDEMDContent string

	// Sentinel-specific
	Watches string

	// Advisor-specific
	AnalysisType    string
	AnalysisContext string
	RunHistory      string

	// Audit-specific
	ChangedFiles string

	// Review-PRD-specific
	TaskSummaries string
}

// ContextOutput is the rendered prompt pair.
type ContextOutput struct {
	Prompt       string
	SystemPrompt string
}

// Execute selects a template by phase, renders it with input data, and returns the prompt.
func (s *ContextService) Execute(_ context.Context, in ContextInput) (ContextOutput, error) {
	tmplStr, systemPrompt, err := templateForPhase(in.Phase)
	if err != nil {
		return ContextOutput{}, err
	}

	data := buildTemplateData(in)

	tmpl, err := template.New(in.Phase).Parse(tmplStr)
	if err != nil {
		return ContextOutput{}, fmt.Errorf("parse template %s: %w", in.Phase, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return ContextOutput{}, fmt.Errorf("execute template %s: %w", in.Phase, err)
	}

	return ContextOutput{
		Prompt:       buf.String(),
		SystemPrompt: systemPrompt,
	}, nil
}

// templateForPhase returns the template string and system prompt for a given phase.
func templateForPhase(phase string) (tmpl string, system string, err error) {
	switch phase {
	case "plan_tasks":
		return constants.PlanTasksTemplate, "You are a software architect.", nil
	case "implement":
		return constants.ImplementTemplate, "You are a TDD software engineer.", nil
	case "review_task":
		return constants.ReviewTaskTemplate, "You are a code reviewer.", nil
	case "review_prd":
		return constants.ReviewPRDTemplate, "You are a PRD alignment reviewer.", nil
	case "audit":
		return constants.AuditTemplate, "You are a convention auditor.", nil
	case "sentinel":
		return constants.SentinelTemplate, "You are a code health sentinel.", nil
	case "advisor":
		return constants.AdvisorTemplate, "You are an optimization advisor.", nil
	case "integration_review":
		return constants.IntegrationReviewTemplate, "You are an integration reviewer.", nil
	default:
		return "", "", fmt.Errorf("unknown phase: %s", phase)
	}
}

// templateData is the flat struct passed to Go templates.
type templateData struct {
	// Common
	ProjectName     string
	Branch          string
	CLAUDEMDContent string

	// PRD fields
	PRDTitle           string
	PRDBody            string
	AcceptanceCriteria string
	DesignDecisions    string

	// Task fields
	TaskTitle string
	TaskSpec  string

	// Implement
	CompletedDeps []DepSummary
	IsRetry       bool
	RetryError    string
	ReviewReasons string

	// Review
	GitDiff string

	// Audit
	ChangedFiles string

	// Review-PRD
	TaskSummaries string

	// Sentinel
	Watches string

	// Advisor
	AnalysisType    string
	AnalysisContext string
	RunHistory      string
}

func buildTemplateData(in ContextInput) templateData {
	return templateData{
		ProjectName:        in.Project.Name,
		Branch:             in.Task.Branch,
		CLAUDEMDContent:    in.CLAUDEMDContent,
		PRDTitle:           in.PRD.Title,
		PRDBody:            in.PRD.Body,
		AcceptanceCriteria: formatAcceptanceCriteria(in),
		DesignDecisions:    formatDesignDecisions(in.PRD.DesignDecisions),
		TaskTitle:          in.Task.Title,
		TaskSpec:           in.Task.Spec,
		CompletedDeps:      in.CompletedDeps,
		IsRetry:            in.IsRetry,
		RetryError:         in.RetryError,
		ReviewReasons:      in.ReviewReasons,
		GitDiff:            in.GitDiff,
		ChangedFiles:       in.ChangedFiles,
		TaskSummaries:      in.TaskSummaries,
		Watches:            in.Watches,
		AnalysisType:       in.AnalysisType,
		AnalysisContext:    in.AnalysisContext,
		RunHistory:         in.RunHistory,
	}
}

// formatAcceptanceCriteria parses JSON acceptance criteria into readable lines.
// Supports both PRD criteria (with verification) and task criteria.
func formatAcceptanceCriteria(in ContextInput) string {
	// Try task criteria first, fall back to PRD criteria.
	src := in.Task.AcceptanceCriteria
	if src == "" {
		src = in.PRD.AcceptanceCriteria
	}

	type criterion struct {
		ID           string `json:"id"`
		Description  string `json:"description"`
		Verification string `json:"verification,omitempty"`
	}

	var criteria []criterion
	if err := json.Unmarshal([]byte(src), &criteria); err != nil {
		// Not JSON — return as-is (plain text criteria).
		return src
	}

	var lines []string
	for _, c := range criteria {
		line := fmt.Sprintf("- [%s] %s", c.ID, c.Description)
		if c.Verification != "" {
			line += fmt.Sprintf(" (verify: %s)", c.Verification)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// formatDesignDecisions parses JSON design decisions into readable lines.
func formatDesignDecisions(raw string) string {
	if raw == "" || raw == "[]" {
		return "None specified."
	}

	var decisions []string
	if err := json.Unmarshal([]byte(raw), &decisions); err != nil {
		return raw
	}

	var lines []string
	for _, d := range decisions {
		lines = append(lines, "- "+d)
	}
	return strings.Join(lines, "\n")
}

func (s *ContextService) Description() string {
	return "Build prompt context for headless agent skills"
}
