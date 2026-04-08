package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/yolo-hq/yolo/core/pkg/claude"
	"github.com/yolo-hq/yolo/core/service"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/constants"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/events"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/helpers"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/lint"
)

// OrchestratorService executes a single task through the full workflow:
// plan -> implement -> test -> audit -> review -> merge.
type OrchestratorService struct {
	service.Base
	Claude     *claude.Client
	Git        *GitService
	Context    *ContextService
	Dependency *DependencyService
	Linter     *LinterService
}

// OrchestratorInput holds the data needed to execute a task workflow.
type OrchestratorInput struct {
	Task    entities.Task
	PRD     entities.PRD
	Project entities.Project
}

// OrchestratorOutput holds the result of a full task execution.
type OrchestratorOutput struct {
	Run          entities.Run
	Steps        []entities.Step
	Review       *entities.Review
	Questions    []entities.Question
	Status       string // "completed" or "failed"
	CostUSD      float64
	CommitHash   string
	FilesChanged []string
	Summary      string
}

// StepParams configures a single step execution.
type StepParams struct {
	RunID      string
	TaskID     string
	Phase      string // plan, implement, test, audit, review
	Skill      string
	Model      string
	WorkDir    string
	Prompt     string
	Config     claude.Config
	IsShell    bool     // true for test step (exec.Command, not agent)
	Commands   []string // shell commands for test step
	ResumeID   string   // session ID to resume (implement step)
}

// StepResult holds the output of a single step.
type StepResult struct {
	Step      entities.Step
	SessionID string
	Output    string
	Failed    bool
	Error     string
	Review    *entities.Review
}

// Execute runs the full task workflow and returns the result.
func (s *OrchestratorService) Execute(ctx context.Context, in OrchestratorInput) (OrchestratorOutput, error) {
	out := OrchestratorOutput{Status: entities.RunFailed}

	// 0. Budget enforcement — check before any work.
	if err := checkBudget(in.Project); err != nil {
		service.EmitEvent(ctx, service.PendingEvent{
			Name: events.BudgetExceeded,
			Data: events.BudgetPayload{
				ProjectName: in.Project.Name,
				Spent:       in.Project.SpentThisMonthUSD,
				Limit:       in.Project.BudgetMonthlyUSD,
				Percentage:  100,
			},
		})
		out.Summary = err.Error()
		return out, nil
	}

	// 0b. Budget warning — log if approaching limit.
	if in.Project.BudgetMonthlyUSD > 0 && in.Project.BudgetWarningAt > 0 {
		ratio := in.Project.SpentThisMonthUSD / in.Project.BudgetMonthlyUSD
		if ratio >= in.Project.BudgetWarningAt {
			service.EmitEvent(ctx, service.PendingEvent{
				Name: events.BudgetWarning,
				Data: events.BudgetPayload{
					ProjectName: in.Project.Name,
					Spent:       in.Project.SpentThisMonthUSD,
					Limit:       in.Project.BudgetMonthlyUSD,
					Percentage:  ratio * 100,
				},
			})
			slog.Warn("budget warning",
				"project", in.Project.Name,
				"spent", in.Project.SpentThisMonthUSD,
				"limit", in.Project.BudgetMonthlyUSD,
				"pct", fmt.Sprintf("%.0f%%", ratio*100),
			)
		}
	}

	// 0c. Emit task started event.
	service.EmitEvent(ctx, service.PendingEvent{
		EntityType: "Task",
		EntityID:   in.Task.ID,
		Name:       events.TaskStarted,
		Data: events.TaskPayload{
			TaskID:      in.Task.ID,
			Title:       in.Task.Title,
			ProjectName: in.Project.Name,
		},
	})

	// 1. Determine working directory.
	workDir := workingDir(in.Project, in.Task.ID)

	// 2. Determine model (with escalation support).
	model := determineModel(in.Task, in.Project)

	// 3. Read CLAUDE.md from working dir.
	claudeMD := readCLAUDEMD(in.Project.LocalPath)

	// 4. Git setup: pull latest, create branch.
	branchName := ""
	if in.Project.UseWorktrees {
		wtPath := filepath.Join(in.Project.LocalPath, ".worktrees", "task-"+in.Task.ID)
		gitOut, err := s.Git.Execute(ctx, GitInput{
			Operation: "worktree_add",
			RepoPath:  in.Project.LocalPath,
			TaskID:    in.Task.ID,
			Path:      wtPath,
		})
		if err != nil {
			return out, fmt.Errorf("create worktree: %w", err)
		}
		branchName = gitOut.BranchName
		workDir = wtPath
	} else {
		if _, err := s.Git.Execute(ctx, GitInput{
			Operation: "pull",
			RepoPath:  in.Project.LocalPath,
			Branch:    in.Task.Branch,
		}); err != nil {
			return out, fmt.Errorf("git pull: %w", err)
		}
		gitOut, err := s.Git.Execute(ctx, GitInput{
			Operation: "branch",
			RepoPath:  in.Project.LocalPath,
			TaskID:    in.Task.ID,
		})
		if err != nil {
			return out, fmt.Errorf("create branch: %w", err)
		}
		branchName = gitOut.BranchName
	}

	// 4b. Run setup commands if defined.
	setupCommands := parseTestCommands(in.Project.SetupCommands)
	if in.Project.SetupCommands != "" && in.Project.SetupCommands != "[]" {
		for _, cmd := range setupCommands {
			if _, err := runShellCommand(ctx, workDir, cmd); err != nil {
				return out, fmt.Errorf("setup command %q: %w", cmd, err)
			}
		}
	}

	// 5. Create Run entity struct.
	runID := ulid.Make().String()
	now := time.Now()
	run := entities.Run{
		TaskID:      in.Task.ID,
		AgentType:   entities.AgentImplementer,
		Status:      entities.RunRunning,
		Model:       model,
		BranchName:  branchName,
		StartedAt:   now,
		SessionName: fmt.Sprintf("factory:task-%s", in.Task.ID),
	}
	run.ID = runID

	// Track escalation.
	if task := in.Task; task.RunCount >= in.Project.EscalationAfterRetries && in.Project.EscalationModel != "" && task.Model == "" {
		run.EscalatedModel = model
	}

	var steps []entities.Step
	var totalCost float64

	// 6. Step 1: PLAN — Opus, read-only.
	planCtx, err := s.Context.Execute(ctx, ContextInput{
		Phase:           "implement",
		Task:            in.Task,
		PRD:             in.PRD,
		Project:         in.Project,
		CLAUDEMDContent: claudeMD,
	})
	if err != nil {
		return out, fmt.Errorf("build plan context: %w", err)
	}

	planResult, err := s.executeStep(ctx, StepParams{
		RunID:  runID,
		TaskID: in.Task.ID,
		Phase:  entities.PhasePlan,
		Skill:  entities.PhasePlan,
		Model:  "opus",
		Config: claude.Config{
			Model:          "opus",
			AllowedTools:   []string{"Read", "Glob", "Grep"},
			Bare:           true,
			BudgetUSD:      1.0,
			PermissionMode: "auto",
			Effort:         "high",
			CWD:            workDir,
			SessionName:    fmt.Sprintf("factory:task-%s:plan", in.Task.ID),
			Timeout:        10 * time.Minute,
		},
		Prompt: planCtx.Prompt,
	})
	if err != nil {
		return out, fmt.Errorf("plan step: %w", err)
	}
	steps = append(steps, planResult.Step)
	totalCost += planResult.Step.CostUSD
	if planResult.Failed {
		return s.buildFailure(ctx, in, run, steps, nil, totalCost, planResult)
	}

	// 7. Step 2: IMPLEMENT — resume plan session.
	implCtx, err := s.Context.Execute(ctx, ContextInput{
		Phase:           "implement",
		Task:            in.Task,
		PRD:             in.PRD,
		Project:         in.Project,
		CLAUDEMDContent: claudeMD,
	})
	if err != nil {
		return out, fmt.Errorf("build implement context: %w", err)
	}

	implBudget := 2.0
	if in.Project.BudgetPerTaskUSD > 0 {
		implBudget = in.Project.BudgetPerTaskUSD
	}

	implResult, err := s.executeStep(ctx, StepParams{
		RunID:    runID,
		TaskID:   in.Task.ID,
		Phase:    entities.PhaseImplement,
		Skill:    entities.PhaseImplement,
		Model:    model,
		ResumeID: planResult.SessionID,
		Config: claude.Config{
			Model:          model,
			AllowedTools:   []string{"Read", "Edit", "Write", "Bash", "Glob", "Grep"},
			Bare:           true,
			BudgetUSD:      implBudget,
			PermissionMode: "auto",
			Effort:         "high",
			CWD:            workDir,
			SessionName:    fmt.Sprintf("factory:task-%s:implement", in.Task.ID),
			Timeout:        15 * time.Minute,
		},
		Prompt: implCtx.Prompt,
	})
	if err != nil {
		return out, fmt.Errorf("implement step: %w", err)
	}
	steps = append(steps, implResult.Step)
	totalCost += implResult.Step.CostUSD
	if implResult.Failed {
		return s.buildFailure(ctx, in, run, steps, nil, totalCost, implResult)
	}

	// 7b. Detect questions in implementation output.
	var questions []entities.Question
	if q := detectQuestion(implResult.Output, in.Task, runID); q != nil {
		questions = append(questions, *q)
	}

	// 8. Step 3: TEST — shell commands, not an agent.
	testCommands := parseTestCommands(in.Project.TestCommands)
	testResult, err := s.executeStep(ctx, StepParams{
		RunID:    runID,
		TaskID:   in.Task.ID,
		Phase:    entities.PhaseTest,
		Skill:    entities.PhaseTest,
		Model:    "shell",
		IsShell:  true,
		Commands: testCommands,
		Config: claude.Config{
			CWD: workDir,
		},
	})
	if err != nil {
		return out, fmt.Errorf("test step: %w", err)
	}
	steps = append(steps, testResult.Step)
	if testResult.Failed {
		return s.buildFailure(ctx, in, run, steps, nil, totalCost, testResult)
	}

	// 9. Step 4: LINT — zero-token static analysis.
	lintResult, err := s.executeLintStep(ctx, in.Task, &run, in.Project, workDir, runID)
	if err != nil {
		return out, fmt.Errorf("lint step: %w", err)
	}
	steps = append(steps, lintResult.Step)
	if lintResult.Failed {
		return s.buildFailure(ctx, in, run, steps, nil, totalCost, lintResult)
	}

	// 10. Step 5: AUDIT — Sonnet, read+bash.
	// Get changed files for audit context.
	diffOut, _ := s.Git.Execute(ctx, GitInput{
		Operation: "diff_files",
		RepoPath:  workDir,
	})
	changedFiles := strings.Join(diffOut.FilesChanged, "\n")

	auditCtx, err := s.Context.Execute(ctx, ContextInput{
		Phase:           "audit",
		Task:            in.Task,
		PRD:             in.PRD,
		Project:         in.Project,
		CLAUDEMDContent: claudeMD,
		ChangedFiles:    changedFiles,
	})
	if err != nil {
		return out, fmt.Errorf("build audit context: %w", err)
	}

	auditResult, err := s.executeStep(ctx, StepParams{
		RunID:  runID,
		TaskID: in.Task.ID,
		Phase:  entities.PhaseAudit,
		Skill:  entities.PhaseAudit,
		Model:  "sonnet",
		Config: claude.Config{
			Model:          "sonnet",
			AllowedTools:   []string{"Read", "Bash", "Glob", "Grep"},
			Bare:           true,
			BudgetUSD:      0.30,
			PermissionMode: "auto",
			Effort:         "medium",
			CWD:            workDir,
			JSONSchema:     constants.AuditSchema,
			SessionName:    fmt.Sprintf("factory:task-%s:audit", in.Task.ID),
			Timeout:        5 * time.Minute,
		},
		Prompt: auditCtx.Prompt,
	})
	if err != nil {
		return out, fmt.Errorf("audit step: %w", err)
	}
	steps = append(steps, auditResult.Step)
	totalCost += auditResult.Step.CostUSD
	if auditResult.Failed {
		return s.buildFailure(ctx, in, run, steps, nil, totalCost, auditResult)
	}

	// 10. Step 5: REVIEW — Sonnet, read-only, NEW session.
	// Get git diff for review context.
	reviewDiffOut, _ := s.Git.Execute(ctx, GitInput{
		Operation: "diff_files",
		RepoPath:  workDir,
	})

	reviewCtx, err := s.Context.Execute(ctx, ContextInput{
		Phase:           "review_task",
		Task:            in.Task,
		PRD:             in.PRD,
		Project:         in.Project,
		CLAUDEMDContent: claudeMD,
		GitDiff:         strings.Join(reviewDiffOut.FilesChanged, "\n"),
	})
	if err != nil {
		return out, fmt.Errorf("build review context: %w", err)
	}

	reviewResult, err := s.executeStep(ctx, StepParams{
		RunID:  runID,
		TaskID: in.Task.ID,
		Phase:  entities.PhaseReview,
		Skill:  entities.PhaseReview,
		Model:  "sonnet",
		Config: claude.Config{
			Model:          "sonnet",
			AllowedTools:   []string{"Read", "Glob", "Grep"},
			Bare:           true,
			BudgetUSD:      0.50,
			PermissionMode: "auto",
			Effort:         "medium",
			CWD:            workDir,
			JSONSchema:     constants.ReviewTaskSchema,
			SessionName:    fmt.Sprintf("factory:task-%s:review", in.Task.ID),
			Timeout:        5 * time.Minute,
		},
		Prompt: reviewCtx.Prompt,
	})
	if err != nil {
		return out, fmt.Errorf("review step: %w", err)
	}
	steps = append(steps, reviewResult.Step)
	totalCost += reviewResult.Step.CostUSD

	// Parse review output into Review entity.
	var review *entities.Review
	if reviewResult.Review != nil {
		review = reviewResult.Review
	}

	if reviewResult.Failed {
		return s.buildFailure(ctx, in, run, steps, review, totalCost, reviewResult)
	}

	// 11. All steps passed — capture files changed BEFORE committing.
	filesOut, _ := s.Git.Execute(ctx, GitInput{
		Operation: "diff_files",
		RepoPath:  workDir,
	})
	filesChanged := filesOut.FilesChanged

	// 12. Git commit.
	summary := helpers.Truncate(implResult.Output, 500)
	commitOut, err := s.Git.Execute(ctx, GitInput{
		Operation: "commit",
		RepoPath:  workDir,
		Message:   fmt.Sprintf("feat: %s\n\nTask: %s", in.Task.Title, in.Task.ID),
	})
	if err != nil {
		return out, fmt.Errorf("git commit: %w", err)
	}
	commitHash := commitOut.CommitHash

	// 13. Merge to target branch + push (if auto_merge).
	if in.Project.AutoMerge {
		if _, err := s.Git.Execute(ctx, GitInput{
			Operation: "merge",
			RepoPath:  in.Project.LocalPath,
			Branch:    in.Task.Branch,
			TaskID:    in.Task.ID,
		}); err != nil {
			return out, fmt.Errorf("git merge: %w", err)
		}

		if _, err := s.Git.Execute(ctx, GitInput{
			Operation: "push",
			RepoPath:  in.Project.LocalPath,
			Branch:    in.Task.Branch,
		}); err != nil {
			return out, fmt.Errorf("git push: %w", err)
		}
	}

	// 14. Git cleanup.
	if in.Project.UseWorktrees {
		wtPath := filepath.Join(in.Project.LocalPath, ".worktrees", "task-"+in.Task.ID)
		if _, err := s.Git.Execute(ctx, GitInput{
			Operation: "worktree_remove",
			RepoPath:  in.Project.LocalPath,
			Path:      wtPath,
		}); err != nil {
			slog.Warn("git worktree cleanup failed", "task_id", in.Task.ID, "error", err)
		}
	}
	if _, err := s.Git.Execute(ctx, GitInput{
		Operation:    "delete_branch",
		RepoPath:     in.Project.LocalPath,
		TaskID:       in.Task.ID,
	}); err != nil {
		slog.Warn("git branch cleanup failed", "task_id", in.Task.ID, "error", err)
	}

	// 15. TODO: Trigger integration review after every N completed tasks.
	// Use ShouldRunIntegrationReview(completedCount, defaultIntegrationReviewEvery)
	// to decide. The job layer should count completed tasks in the PRD and call
	// IntegrationReviewService.Execute with the combined diff from recent tasks.

	// 16. Emit task completed event.
	service.EmitEvent(ctx, service.PendingEvent{
		EntityType: "Task",
		EntityID:   in.Task.ID,
		Name:       events.TaskCompleted,
		Data: events.TaskPayload{
			TaskID:      in.Task.ID,
			Title:       in.Task.Title,
			ProjectName: in.Project.Name,
			CostUSD:     totalCost,
		},
	})

	// 17. Build final output.
	completedAt := time.Now()
	run.Status = entities.RunCompleted
	run.CostUSD = totalCost
	run.CommitHash = commitHash
	run.FilesChanged = helpers.ToJSON(filesChanged)
	run.CompletedAt = &completedAt
	run.Result = summary

	return OrchestratorOutput{
		Run:          run,
		Steps:        steps,
		Review:       review,
		Questions:    questions,
		Status:       entities.RunCompleted,
		CostUSD:      totalCost,
		CommitHash:   commitHash,
		FilesChanged: filesChanged,
		Summary:      summary,
	}, nil
}

// executeStep runs a single step (agent or shell) and returns the result.
func (s *OrchestratorService) executeStep(ctx context.Context, params StepParams) (*StepResult, error) {
	stepID := ulid.Make().String()
	startedAt := time.Now()

	step := entities.Step{
		RunID:   params.RunID,
		Phase:   params.Phase,
		Skill:   params.Skill,
		Status:  entities.StepRunning,
		Model:   params.Model,
		StartedAt: startedAt,
	}
	step.ID = stepID

	result := &StepResult{Step: step}

	if params.IsShell {
		// Test step: run shell commands.
		var combinedOutput strings.Builder
		for _, cmd := range params.Commands {
			cmdOut, err := runShellCommand(ctx, params.Config.CWD, cmd)
			combinedOutput.WriteString(fmt.Sprintf("$ %s\n%s\n", cmd, cmdOut))
			if err != nil {
				completedAt := time.Now()
				step.Status = entities.StepFailed
				step.CompletedAt = &completedAt
				step.DurationMs = int(completedAt.Sub(startedAt).Milliseconds())
				step.OutputSummary = helpers.Truncate(combinedOutput.String(), 500)
				result.Step = step
				result.Failed = true
				result.Error = fmt.Sprintf("command failed: %s: %s", cmd, cmdOut)
				result.Output = combinedOutput.String()
				return result, nil
			}
		}
		completedAt := time.Now()
		step.Status = entities.StepCompleted
		step.CompletedAt = &completedAt
		step.DurationMs = int(completedAt.Sub(startedAt).Milliseconds())
		step.OutputSummary = helpers.Truncate(combinedOutput.String(), 500)
		result.Step = step
		result.Output = combinedOutput.String()
		return result, nil
	}

	// Agent step: run or resume Claude.
	var claudeResult *claude.Result
	var err error

	if params.ResumeID != "" {
		claudeResult, err = s.Claude.Resume(ctx, params.ResumeID, params.Config, params.Prompt)
	} else {
		claudeResult, err = s.Claude.Run(ctx, params.Config, params.Prompt)
	}

	completedAt := time.Now()
	step.DurationMs = int(completedAt.Sub(startedAt).Milliseconds())
	step.CompletedAt = &completedAt

	if err != nil {
		step.Status = entities.StepFailed
		step.OutputSummary = helpers.Truncate(err.Error(), 500)
		result.Step = step
		result.Failed = true
		result.Error = err.Error()
		return result, nil
	}

	// Populate step metrics from claude result.
	step.SessionID = claudeResult.SessionID
	step.CostUSD = claudeResult.CostUSD
	step.TokensIn = claudeResult.Usage.InputTokens
	step.TokensOut = claudeResult.Usage.OutputTokens
	result.SessionID = claudeResult.SessionID
	result.Output = claudeResult.Text

	if claudeResult.IsError {
		step.Status = entities.StepFailed
		step.OutputSummary = helpers.Truncate(claudeResult.Text, 500)
		result.Step = step
		result.Failed = true
		result.Error = claudeResult.Text
		return result, nil
	}

	// Parse structured output for audit and review steps.
	switch params.Phase {
	case entities.PhaseAudit:
		failed, errMsg := parseAuditOutput(claudeResult.StructuredOutput)
		if failed {
			step.Status = entities.StepFailed
			step.OutputSummary = helpers.Truncate(errMsg, 500)
			result.Step = step
			result.Failed = true
			result.Error = errMsg
			return result, nil
		}
	case entities.PhaseReview:
		review, failed, errMsg := parseReviewOutput(claudeResult.StructuredOutput, params.RunID, params.TaskID, claudeResult)
		result.Review = review
		if failed {
			step.Status = entities.StepFailed
			step.OutputSummary = helpers.Truncate(errMsg, 500)
			result.Step = step
			result.Failed = true
			result.Error = errMsg
			return result, nil
		}
	}

	step.Status = entities.StepCompleted
	step.OutputSummary = helpers.Truncate(claudeResult.Text, 500)
	result.Step = step
	return result, nil
}

// buildFailure constructs an OrchestratorOutput for early-return (failure) cases.
// It also emits TaskFailed event and pushes the branch if push_failed_branches is set.
func (s *OrchestratorService) buildFailure(ctx context.Context, in OrchestratorInput, run entities.Run, steps []entities.Step, review *entities.Review, totalCost float64, lastResult *StepResult) (OrchestratorOutput, error) {
	completedAt := time.Now()
	run.Status = entities.RunFailed
	run.CostUSD = totalCost
	run.Error = lastResult.Error
	run.CompletedAt = &completedAt

	service.EmitEvent(ctx, service.PendingEvent{
		EntityType: "Task",
		EntityID:   run.TaskID,
		Name:       events.TaskFailed,
		Data: events.TaskPayload{
			TaskID:      run.TaskID,
			Title:       in.Task.Title,
			ProjectName: in.Project.Name,
			CostUSD:     totalCost,
			Error:       lastResult.Error,
		},
	})

	if in.Project.PushFailedBranches && run.BranchName != "" {
		if _, err := s.Git.Execute(ctx, GitInput{
			Operation: "push",
			RepoPath:  workingDir(in.Project, in.Task.ID),
			Branch:    run.BranchName,
		}); err != nil {
			slog.Warn("push failed branch failed", "task_id", in.Task.ID, "branch", run.BranchName, "error", err)
		}
	}

	return OrchestratorOutput{
		Run:     run,
		Steps:   steps,
		Review:  review,
		Status:  entities.RunFailed,
		CostUSD: totalCost,
		Summary: lastResult.Error,
	}, nil
}

// workingDir returns the directory to run agents in.
func workingDir(project entities.Project, taskID string) string {
	if project.UseWorktrees {
		return filepath.Join(project.LocalPath, ".worktrees", "task-"+taskID)
	}
	return project.LocalPath
}

// determineModel returns the model to use:
// task override > escalation (after retries) > project default > "sonnet".
func determineModel(task entities.Task, project entities.Project) string {
	if task.Model != "" {
		return task.Model
	}
	if task.RunCount >= project.EscalationAfterRetries && project.EscalationModel != "" {
		return project.EscalationModel
	}
	if project.DefaultModel != "" {
		return project.DefaultModel
	}
	return "sonnet"
}

// checkBudget verifies the project has not exceeded its monthly budget.
func checkBudget(project entities.Project) error {
	if project.BudgetMonthlyUSD > 0 && project.SpentThisMonthUSD >= project.BudgetMonthlyUSD {
		return fmt.Errorf("monthly budget exceeded: spent $%.2f of $%.2f limit",
			project.SpentThisMonthUSD, project.BudgetMonthlyUSD)
	}
	return nil
}

// detectQuestion scans agent output for "QUESTION:" prefix and extracts the question.
func detectQuestion(resultText string, task entities.Task, runID string) *entities.Question {
	upper := strings.ToUpper(resultText)
	idx := strings.Index(upper, "QUESTION:")
	if idx == -1 {
		return nil
	}
	questionText := strings.TrimSpace(resultText[idx+9:])
	if nl := strings.Index(questionText, "\n"); nl > 0 {
		questionText = questionText[:nl]
	}
	if questionText == "" {
		return nil
	}

	q := &entities.Question{
		TaskID:     task.ID,
		RunID:      runID,
		Body:       questionText,
		Context:    "Detected during implementation step",
		Confidence: entities.ConfidenceMedium,
		Status:     entities.QuestionOpen,
	}
	q.ID = ulid.Make().String()
	return q
}

// parseTestCommands parses the project's test_commands JSON array.
func parseTestCommands(raw string) []string {
	if raw == "" || raw == "[]" {
		return []string{"go build ./...", "go test ./..."}
	}
	var cmds []string
	if err := json.Unmarshal([]byte(raw), &cmds); err != nil {
		return []string{"go build ./...", "go test ./..."}
	}
	if len(cmds) == 0 {
		return []string{"go build ./...", "go test ./..."}
	}
	return cmds
}

// shellMetaChars are characters that indicate shell injection attempts.
var shellMetaChars = []string{"|", ";", "&", "$", "`", "(", ")", "{", "}", "<", ">", "!", "~"}

// validateCommand checks that a command string has no shell metacharacters.
func validateCommand(command string) error {
	for _, ch := range shellMetaChars {
		if strings.Contains(command, ch) {
			return fmt.Errorf("command contains shell metacharacter %q: %s", ch, command)
		}
	}
	return nil
}

// runShellCommand executes a command by splitting on whitespace (no shell).
func runShellCommand(ctx context.Context, dir string, command string) (string, error) {
	if err := validateCommand(command); err != nil {
		return "", err
	}

	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return stdout.String() + stderr.String(), err
	}
	return stdout.String(), nil
}

// auditOutput is the structured output from the audit agent.
type auditOutput struct {
	Passed     bool     `json:"passed"`
	Violations []string `json:"violations"`
	Warnings   []string `json:"warnings"`
}

// parseAuditOutput parses audit structured output, returns (failed, errorMsg).
func parseAuditOutput(raw json.RawMessage) (bool, string) {
	if len(raw) == 0 {
		return false, ""
	}
	var out auditOutput
	if err := json.Unmarshal(raw, &out); err != nil {
		return true, fmt.Sprintf("failed to parse audit output: %s", err)
	}
	if !out.Passed {
		return true, fmt.Sprintf("audit failed: %s", strings.Join(out.Violations, "; "))
	}
	return false, ""
}

// reviewOutput is the structured output from the review agent.
type reviewOutput struct {
	Verdict         string   `json:"verdict"`
	CriteriaResults json.RawMessage `json:"criteria_results"`
	AntiPatterns    []string `json:"anti_patterns"`
	Reasons         []string `json:"reasons"`
	Suggestions     []string `json:"suggestions"`
}

// parseReviewOutput parses review structured output into a Review entity.
func parseReviewOutput(raw json.RawMessage, runID, taskID string, result *claude.Result) (*entities.Review, bool, string) {
	if len(raw) == 0 {
		return nil, false, ""
	}
	var out reviewOutput
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, true, fmt.Sprintf("failed to parse review output: %s", err)
	}

	review := &entities.Review{
		RunID:           runID,
		TaskID:          taskID,
		SessionID:       result.SessionID,
		Model:           "sonnet",
		Verdict:         out.Verdict,
		Reasons:         helpers.ToJSON(out.Reasons),
		AntiPatterns:    helpers.ToJSON(out.AntiPatterns),
		CriteriaResults: string(out.CriteriaResults),
		Suggestions:     helpers.ToJSON(out.Suggestions),
		CostUSD:         result.CostUSD,
	}
	review.ID = ulid.Make().String()

	if out.Verdict == entities.ReviewFail {
		return review, true, fmt.Sprintf("review failed: %s", strings.Join(out.Reasons, "; "))
	}
	return review, false, ""
}

// executeLintStep runs zero-token static analysis checks on the project code.
func (s *OrchestratorService) executeLintStep(ctx context.Context, task entities.Task, run *entities.Run, project entities.Project, workDir string, runID string) (*StepResult, error) {
	stepID := ulid.Make().String()
	startedAt := time.Now()

	step := entities.Step{
		RunID:     runID,
		Phase:     entities.PhaseLint,
		Skill:     "factory-lint",
		Status:    entities.StepRunning,
		Model:     "",
		StartedAt: startedAt,
	}
	step.ID = stepID

	// Get changed files for targeted lint.
	diffOut, _ := s.Git.Execute(ctx, GitInput{
		Operation: "diff_files",
		RepoPath:  workDir,
	})

	lintOut, err := s.Linter.Execute(ctx, LinterInput{
		ProjectPath:  workDir,
		ChangedFiles: diffOut.FilesChanged,
	})
	if err != nil {
		completedAt := time.Now()
		step.Status = entities.StepFailed
		step.CompletedAt = &completedAt
		step.DurationMs = int(completedAt.Sub(startedAt).Milliseconds())
		step.OutputSummary = helpers.Truncate(err.Error(), 500)
		return &StepResult{Step: step, Failed: true, Error: err.Error()}, nil
	}

	completedAt := time.Now()
	step.CompletedAt = &completedAt
	step.DurationMs = int(completedAt.Sub(startedAt).Milliseconds())

	summary := fmt.Sprintf("checks=%d passed=%d failed=%d findings=%d",
		lintOut.ChecksRun, lintOut.ChecksPassed, lintOut.ChecksFailed, len(lintOut.Findings))

	if !lintOut.Passed {
		// Build error text from findings.
		var errParts []string
		for _, f := range lintOut.Findings {
			if f.Severity == lint.SeverityError {
				errParts = append(errParts, fmt.Sprintf("%s:%d: %s", f.File, f.Line, f.Message))
			}
		}
		errText := strings.Join(errParts, "\n")
		step.Status = entities.StepFailed
		step.OutputSummary = helpers.Truncate(summary+"\n"+errText, 500)
		return &StepResult{Step: step, Failed: true, Error: errText, Output: summary}, nil
	}

	step.Status = entities.StepCompleted
	step.OutputSummary = summary
	return &StepResult{Step: step, Output: summary}, nil
}

func (s *OrchestratorService) Description() string { return "Execute full implementation workflow for a task" }
