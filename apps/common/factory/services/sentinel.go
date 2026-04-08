package services

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/oklog/ulid/v2"
	"github.com/yolo-hq/yolo/core/pkg/claude"
	"github.com/yolo-hq/yolo/core/service"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/events"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/helpers"
)

// SentinelService runs health checks against a project and produces findings.
type SentinelService struct {
	service.Base
	Claude *claude.Client
}

// SentinelInput specifies which watches to run against a project.
type SentinelInput struct {
	Project entities.Project
	Watches []string // "build_health", "test_health", "security", "convention_drift", "orphaned_runs"
}

// Finding is a single health check result.
type Finding struct {
	Watch    string
	Severity string // "critical", "warning", "info"
	Message  string
	Action   string // "create_task", "create_suggestion", "none"
}

// SentinelOutput holds all findings and entities to create.
type SentinelOutput struct {
	Findings            []Finding
	TasksToCreate       []entities.Task
	SuggestionsToCreate []entities.Suggestion
}

// Execute runs each watch and collects findings.
func (s *SentinelService) Execute(ctx context.Context, in SentinelInput) (SentinelOutput, error) {
	var out SentinelOutput

	for _, watch := range in.Watches {
		findings, err := s.runWatch(ctx, watch, in.Project)
		if err != nil {
			// Non-fatal: record as info finding.
			out.Findings = append(out.Findings, Finding{
				Watch:    watch,
				Severity: "info",
				Message:  fmt.Sprintf("watch error: %s", err),
				Action:   "none",
			})
			continue
		}
		out.Findings = append(out.Findings, findings...)
	}

	// Emit events for critical findings.
	for _, f := range out.Findings {
		switch {
		case f.Watch == "build_health" && f.Severity == "critical":
			service.EmitEvent(ctx, service.PendingEvent{
				Name: events.SentinelBuildBrokenName,
				Data: events.SentinelPayload{
					ProjectID: in.Project.ID,
					Error:       f.Message,
					Severity:    f.Severity,
				},
			})
		case f.Watch == "security" && f.Severity == "critical":
			service.EmitEvent(ctx, service.PendingEvent{
				Name: events.SentinelSecurityVulnName,
				Data: events.SentinelPayload{
					ProjectID: in.Project.ID,
					Error:       f.Message,
					Severity:    f.Severity,
				},
			})
		}
	}

	// Convert findings to tasks/suggestions.
	for _, f := range out.Findings {
		switch f.Action {
		case "create_task":
			task := entities.Task{
				ProjectID: in.Project.ID,
				Title:     fmt.Sprintf("[sentinel] %s", f.Message),
				Spec:      f.Message,
				Status:    entities.TaskQueued,
				Branch:    in.Project.DefaultBranch,
			}
			task.ID = ulid.Make().String()
			out.TasksToCreate = append(out.TasksToCreate, task)
		case "create_suggestion":
			sug := entities.Suggestion{
				ProjectID: in.Project.ID,
				Source:    "sentinel",
				Category:  categoryFromWatch(f.Watch),
				Title:     fmt.Sprintf("[sentinel] %s", helpers.Truncate(f.Message, 80)),
				Body:      f.Message,
				Priority:  f.Severity,
			}
			sug.ID = ulid.Make().String()
			out.SuggestionsToCreate = append(out.SuggestionsToCreate, sug)
		}
	}

	return out, nil
}

// runWatch executes a single watch type and returns findings.
func (s *SentinelService) runWatch(ctx context.Context, watch string, project entities.Project) ([]Finding, error) {
	switch watch {
	case "build_health":
		return s.checkBuild(ctx, project)
	case "test_health":
		return s.checkTests(ctx, project)
	case "security":
		return s.checkSecurity(ctx, project)
	case "convention_drift":
		return s.checkConventions(ctx, project)
	default:
		return []Finding{{
			Watch:    watch,
			Severity: "info",
			Message:  fmt.Sprintf("unknown watch: %s", watch),
			Action:   "none",
		}}, nil
	}
}

func (s *SentinelService) checkBuild(ctx context.Context, project entities.Project) ([]Finding, error) {
	output, err := runCmd(ctx, project.LocalPath, "go", "build", "./...")
	if err != nil {
		return []Finding{{
			Watch:    "build_health",
			Severity: "critical",
			Message:  fmt.Sprintf("build failed: %s", helpers.Truncate(output, 300)),
			Action:   "create_task",
		}}, nil
	}
	return []Finding{{
		Watch:    "build_health",
		Severity: "info",
		Message:  "build passing",
		Action:   "none",
	}}, nil
}

func (s *SentinelService) checkTests(ctx context.Context, project entities.Project) ([]Finding, error) {
	output, err := runCmd(ctx, project.LocalPath, "go", "test", "./...")
	if err != nil {
		return []Finding{{
			Watch:    "test_health",
			Severity: "critical",
			Message:  fmt.Sprintf("tests failed: %s", helpers.Truncate(output, 300)),
			Action:   "create_task",
		}}, nil
	}
	return []Finding{{
		Watch:    "test_health",
		Severity: "info",
		Message:  "tests passing",
		Action:   "none",
	}}, nil
}

func (s *SentinelService) checkSecurity(ctx context.Context, project entities.Project) ([]Finding, error) {
	// Best-effort: skip if govulncheck not installed.
	if _, err := exec.LookPath("govulncheck"); err != nil {
		return []Finding{{
			Watch:    "security",
			Severity: "info",
			Message:  "govulncheck not installed, skipping",
			Action:   "none",
		}}, nil
	}

	output, err := runCmd(ctx, project.LocalPath, "govulncheck", "./...")
	if err != nil {
		return []Finding{{
			Watch:    "security",
			Severity: "critical",
			Message:  fmt.Sprintf("vulnerabilities found: %s", helpers.Truncate(output, 300)),
			Action:   "create_task",
		}}, nil
	}
	return []Finding{{
		Watch:    "security",
		Severity: "info",
		Message:  "no vulnerabilities found",
		Action:   "none",
	}}, nil
}

func (s *SentinelService) checkConventions(ctx context.Context, project entities.Project) ([]Finding, error) {
	output, err := runCmd(ctx, project.LocalPath, "yolo", "audit")
	if err != nil {
		return []Finding{{
			Watch:    "convention_drift",
			Severity: "warning",
			Message:  fmt.Sprintf("convention violations: %s", helpers.Truncate(output, 300)),
			Action:   "create_suggestion",
		}}, nil
	}
	return []Finding{{
		Watch:    "convention_drift",
		Severity: "info",
		Message:  "conventions passing",
		Action:   "none",
	}}, nil
}

// categoryFromWatch maps watch types to suggestion categories.
func categoryFromWatch(watch string) string {
	switch watch {
	case "build_health", "test_health":
		return "bug"
	case "security":
		return "bug"
	case "convention_drift":
		return "refactor"
	default:
		return "refactor"
	}
}

// runCmd executes a command and returns combined output.
func runCmd(ctx context.Context, dir string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (s *SentinelService) Description() string { return "Monitor project health and detect issues" }
