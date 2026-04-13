package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	yolostrings "github.com/yolo-hq/yolo/core/strings"
	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/pkg/claude"
	"github.com/yolo-hq/yolo/core/read"
	"github.com/yolo-hq/yolo/core/service"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/events"
)

// OrphanedRunThreshold is how long a run may stay in "running" before it's orphaned.
const OrphanedRunThreshold = 2 * time.Hour

// SentinelService runs health checks against a project and produces findings.
type SentinelService struct {
	service.Base
	Claude          *claude.Client
	TaskWrite       entity.WriteRepository[entities.Task]
	SuggestionWrite entity.WriteRepository[entities.Suggestion]
}

// SentinelInput specifies which watches to run against a project.
type SentinelInput struct {
	ProjectID string
	Watches   []string // "build_health", "test_health", "security", "convention_drift", "orphaned_runs"
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

// Execute loads the project, runs each watch, persists tasks/suggestions, and returns findings.
func (s *SentinelService) Execute(ctx context.Context, in SentinelInput) (SentinelOutput, error) {
	var out SentinelOutput

	// Load project.
	project, err := read.FindOne[entities.Project](ctx, in.ProjectID)
	if err != nil {
		return out, fmt.Errorf("load project: %w", err)
	}
	if project.ID == "" {
		return out, fmt.Errorf("project %s not found", in.ProjectID)
	}

	for _, watch := range in.Watches {
		findings, err := s.runWatch(ctx, watch, project)
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
					ProjectID: project.ID,
					Error:     f.Message,
					Severity:  f.Severity,
				},
			})
		case f.Watch == "security" && f.Severity == "critical":
			service.EmitEvent(ctx, service.PendingEvent{
				Name: events.SentinelSecurityVulnName,
				Data: events.SentinelPayload{
					ProjectID: project.ID,
					Error:     f.Message,
					Severity:  f.Severity,
				},
			})
		}
	}

	// Convert findings to tasks/suggestions and persist.
	for _, f := range out.Findings {
		switch f.Action {
		case "create_task":
			task := entities.Task{
				ProjectID: project.ID,
				Title:     fmt.Sprintf("[sentinel] %s", f.Message),
				Spec:      f.Message,
				Status:    string(enums.TaskStatusQueued),
				Branch:    project.DefaultBranch,
			}
			task.ID = ulid.Make().String()
			if _, err := s.TaskWrite.Insert(ctx, &task); err != nil {
				return out, fmt.Errorf("insert task: %w", err)
			}
			out.TasksToCreate = append(out.TasksToCreate, task)
		case "create_suggestion":
			sug := entities.Suggestion{
				ProjectID: project.ID,
				Source:    "sentinel",
				Category:  categoryFromWatch(f.Watch),
				Title:     fmt.Sprintf("[sentinel] %s", yolostrings.Truncate(f.Message, 80)),
				Body:      f.Message,
				Priority:  f.Severity,
			}
			sug.ID = ulid.Make().String()
			if _, err := s.SuggestionWrite.Insert(ctx, &sug); err != nil {
				return out, fmt.Errorf("insert suggestion: %w", err)
			}
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
	case "orphaned_runs":
		return s.checkOrphanedRuns(ctx, project)
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
			Message:  fmt.Sprintf("build failed: %s", yolostrings.Truncate(output, 300)),
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
			Message:  fmt.Sprintf("tests failed: %s", yolostrings.Truncate(output, 300)),
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
	var findings []Finding

	// 1. go vet — standard static analysis.
	if output, err := runCmd(ctx, project.LocalPath, "go", "vet", "./..."); err != nil {
		findings = append(findings, Finding{
			Watch:    "security",
			Severity: "warning",
			Message:  fmt.Sprintf("go vet issues: %s", yolostrings.Truncate(output, 300)),
			Action:   "create_suggestion",
		})
	} else {
		findings = append(findings, Finding{
			Watch:    "security",
			Severity: "info",
			Message:  "go vet clean",
			Action:   "none",
		})
	}

	// 2. govulncheck — skip if not installed.
	if _, err := exec.LookPath("govulncheck"); err != nil {
		findings = append(findings, Finding{
			Watch:    "security",
			Severity: "info",
			Message:  "govulncheck not installed, skipping",
			Action:   "none",
		})
	} else if output, err := runCmd(ctx, project.LocalPath, "govulncheck", "./..."); err != nil {
		findings = append(findings, Finding{
			Watch:    "security",
			Severity: "warning",
			Message:  fmt.Sprintf("vulnerabilities found: %s", yolostrings.Truncate(output, 300)),
			Action:   "create_suggestion",
		})
	} else {
		findings = append(findings, Finding{
			Watch:    "security",
			Severity: "info",
			Message:  "no vulnerabilities found",
			Action:   "none",
		})
	}

	// 3. Scan for hardcoded credentials in Go source (excluding test files).
	credFindings := scanHardcodedCredentials(project.LocalPath)
	findings = append(findings, credFindings...)

	return findings, nil
}

// credPattern matches Go string literals containing credential-like keys.
var credPattern = regexp.MustCompile(`(?i)(password|secret|api_key)\s*=\s*"[^"]+"`)

// scanHardcodedCredentials walks the project looking for hardcoded credentials in non-test Go files.
func scanHardcodedCredentials(root string) []Finding {
	var findings []Finding

	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}

		matches := credPattern.FindAllString(string(data), -1)
		for _, m := range matches {
			findings = append(findings, Finding{
				Watch:    "security",
				Severity: "warning",
				Message:  fmt.Sprintf("possible hardcoded credential in %s: %s", path, yolostrings.Truncate(m, 100)),
				Action:   "create_suggestion",
			})
		}
		return nil
	})

	return findings
}

func (s *SentinelService) checkConventions(ctx context.Context, project entities.Project) ([]Finding, error) {
	output, err := runCmd(ctx, project.LocalPath, "yolo", "audit")
	if err != nil {
		return []Finding{{
			Watch:    "convention_drift",
			Severity: "warning",
			Message:  fmt.Sprintf("convention violations: %s", yolostrings.Truncate(output, 300)),
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

// checkOrphanedRuns finds runs stuck in "running" status past OrphanedRunThreshold.
func (s *SentinelService) checkOrphanedRuns(ctx context.Context, project entities.Project) ([]Finding, error) {
	cutoff := time.Now().Add(-OrphanedRunThreshold)
	runs, err := read.FindMany[entities.Run](ctx,
		read.Eq(fields.Run.Status.Name(), string(enums.RunStatusRunning)),
		read.Limit(1000),
	)
	if err != nil {
		return nil, fmt.Errorf("query runs: %w", err)
	}

	var findings []Finding
	for _, r := range runs {
		if r.StartedAt.Before(cutoff) {
			findings = append(findings, Finding{
				Watch:    "orphaned_runs",
				Severity: "warning",
				Message:  fmt.Sprintf("run %s stuck in running since %s (task %s)", r.ID, r.StartedAt.Format(time.RFC3339), r.TaskID),
				Action:   "create_suggestion",
			})
		}
	}

	if len(findings) == 0 {
		findings = append(findings, Finding{
			Watch:    "orphaned_runs",
			Severity: "info",
			Message:  "no orphaned runs",
			Action:   "none",
		})
	}

	return findings, nil
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
	case "orphaned_runs":
		return "bug"
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
