package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/constants"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/helpers/lint"
	"github.com/yolo-hq/yolo/core/pkg/claude"
	yolostrings "github.com/yolo-hq/yolo/core/strings"
)

// runAllSteps executes steps 6-10 (plan, implement, test, lint, audit, review).
// On step failure/escalation it appends the step to env.steps and returns early so
// Execute can route correctly. It never returns a hard error for step-level failures.
//
// Returns the per-step results (for review extraction), detected questions,
// and any hard error.
func (s *OrchestratorService) runAllSteps(ctx context.Context, env *orchEnv) ([]*StepResult, []entities.Question, error) {
	runID := env.run.ID

	// 6. Step 1: PLAN — Opus, read-only.
	planCtx, err := s.Context.Execute(ctx, ContextInput{
		Phase:           "implement",
		Task:            env.task,
		PRD:             env.prd,
		Project:         env.project,
		CLAUDEMDContent: env.claudeMD,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("build plan context: %w", err)
	}

	planCfg := phaseConfig("plan")
	planCfg.CWD = env.workDir
	planCfg.SessionName = fmt.Sprintf("factory:task-%s:plan", env.task.ID)
	planResult, err := s.executeStep(ctx, StepParams{
		RunID:  runID,
		TaskID: env.task.ID,
		Phase:  string(enums.StepPhasePlan),
		Skill:  string(enums.StepPhasePlan),
		Model:  planCfg.Model,
		Config: planCfg,
		Prompt: planCtx.Prompt,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("plan step: %w", err)
	}
	env.steps = append(env.steps, planResult.Step)
	env.totalCost += planResult.Step.CostUSD
	if !stepCanContinue(planResult) {
		return []*StepResult{planResult}, nil, nil
	}
	logConcerns(planResult)

	// 7. Step 2: IMPLEMENT — resume plan session.
	implCtx, err := s.Context.Execute(ctx, ContextInput{
		Phase:           "implement",
		Task:            env.task,
		PRD:             env.prd,
		Project:         env.project,
		CLAUDEMDContent: env.claudeMD,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("build implement context: %w", err)
	}

	implCfg := phaseConfig("implement")
	implCfg.Model = env.model
	implCfg.CWD = env.workDir
	implCfg.SessionName = fmt.Sprintf("factory:task-%s:implement", env.task.ID)
	if env.project.BudgetPerTaskUSD > 0 {
		implCfg.BudgetUSD = env.project.BudgetPerTaskUSD
	}

	// Compact the plan transcript before handing off to impl.
	planSummary := s.compactPlanOutput(ctx, env.workDir, planResult.Output)
	implPrompt := withPlanSummary(implCtx.Prompt, planSummary)

	implResult, err := s.executeStep(ctx, StepParams{
		RunID:  runID,
		TaskID: env.task.ID,
		Phase:  string(enums.StepPhaseImplement),
		Skill:  string(enums.StepPhaseImplement),
		Model:  env.model,
		Config: implCfg,
		Prompt: implPrompt,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("implement step: %w", err)
	}
	env.steps = append(env.steps, implResult.Step)
	env.totalCost += implResult.Step.CostUSD
	if !stepCanContinue(implResult) {
		return []*StepResult{implResult}, nil, nil
	}
	logConcerns(implResult)

	// 7b. Detect questions in implementation output.
	var questions []entities.Question
	if q := detectQuestion(implResult.Output, env.task, runID); q != nil {
		questions = append(questions, *q)
	}

	// 8. Step 3: TEST — shell commands, not an agent.
	testCommands := parseTestCommands(env.project.TestCommands)
	testResult, err := s.executeStep(ctx, StepParams{
		RunID:    runID,
		TaskID:   env.task.ID,
		Phase:    string(enums.StepPhaseTest),
		Skill:    string(enums.StepPhaseTest),
		Model:    "shell",
		IsShell:  true,
		Commands: testCommands,
		Config: claude.Config{
			CWD: env.workDir,
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("test step: %w", err)
	}
	env.steps = append(env.steps, testResult.Step)
	if !stepCanContinue(testResult) {
		return []*StepResult{testResult}, questions, nil
	}

	// 9. Step 4: LINT — zero-token static analysis.
	lintResult, err := s.executeLintStep(ctx, env.task, &env.run, env.project, env.workDir, runID)
	if err != nil {
		return nil, nil, fmt.Errorf("lint step: %w", err)
	}
	env.steps = append(env.steps, lintResult.Step)
	if !stepCanContinue(lintResult) {
		return []*StepResult{lintResult}, questions, nil
	}

	// 10. Step 5: AUDIT — Sonnet, read+bash.
	diffOut, _ := s.Git.Execute(ctx, GitInput{
		Operation: "diff_files",
		RepoPath:  env.workDir,
	})
	changedFiles := strings.Join(diffOut.FilesChanged, "\n")

	auditCtx, err := s.Context.Execute(ctx, ContextInput{
		Phase:           "audit",
		Task:            env.task,
		PRD:             env.prd,
		Project:         env.project,
		CLAUDEMDContent: env.claudeMD,
		ChangedFiles:    changedFiles,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("build audit context: %w", err)
	}

	auditCfg := phaseConfig("audit")
	auditCfg.CWD = env.workDir
	auditCfg.JSONSchema = constants.AuditSchema
	auditCfg.SessionName = fmt.Sprintf("factory:task-%s:audit", env.task.ID)
	auditResult, err := s.executeStep(ctx, StepParams{
		RunID:  runID,
		TaskID: env.task.ID,
		Phase:  string(enums.StepPhaseAudit),
		Skill:  string(enums.StepPhaseAudit),
		Model:  auditCfg.Model,
		Config: auditCfg,
		Prompt: auditCtx.Prompt,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("audit step: %w", err)
	}
	env.steps = append(env.steps, auditResult.Step)
	env.totalCost += auditResult.Step.CostUSD
	if !stepCanContinue(auditResult) {
		return []*StepResult{auditResult}, questions, nil
	}

	// 11. Step 6: REVIEW — Sonnet, read-only, NEW session.
	reviewDiffOut, _ := s.Git.Execute(ctx, GitInput{
		Operation: "diff_files",
		RepoPath:  env.workDir,
	})

	reviewCtx, err := s.Context.Execute(ctx, ContextInput{
		Phase:           "review_task",
		Task:            env.task,
		PRD:             env.prd,
		Project:         env.project,
		CLAUDEMDContent: env.claudeMD,
		GitDiff:         strings.Join(reviewDiffOut.FilesChanged, "\n"),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("build review context: %w", err)
	}

	reviewCfg := phaseConfig("review")
	reviewCfg.CWD = env.workDir
	reviewCfg.JSONSchema = constants.ReviewTaskSchema
	reviewCfg.SessionName = fmt.Sprintf("factory:task-%s:review", env.task.ID)
	reviewResult, err := s.executeStep(ctx, StepParams{
		RunID:  runID,
		TaskID: env.task.ID,
		Phase:  string(enums.StepPhaseReview),
		Skill:  string(enums.StepPhaseReview),
		Model:  reviewCfg.Model,
		Config: reviewCfg,
		Prompt: reviewCtx.Prompt,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("review step: %w", err)
	}
	env.steps = append(env.steps, reviewResult.Step)
	env.totalCost += reviewResult.Step.CostUSD

	return []*StepResult{planResult, implResult, testResult, lintResult, auditResult, reviewResult}, questions, nil
}

// stepCanContinue returns true when it is safe to proceed to the next step.
// done and done_with_concerns both allow continuation; all other statuses stop the pipeline.
func stepCanContinue(r *StepResult) bool {
	return r.Status == constants.StepResultDone || r.Status == constants.StepResultDoneWithConcerns
}

// logConcerns emits a warning when a step completed with concerns.
func logConcerns(r *StepResult) {
	if r.Status == constants.StepResultDoneWithConcerns && r.Concerns != "" {
		slog.Warn("step completed with concerns",
			"phase", r.Step.Phase,
			"concerns", r.Concerns,
		)
	}
}

// executeStep runs a single step (agent or shell) and returns the result.
func (s *OrchestratorService) executeStep(ctx context.Context, params StepParams) (*StepResult, error) {
	stepID := ulid.Make().String()
	startedAt := time.Now()

	step := entities.Step{
		RunID:     params.RunID,
		Phase:     params.Phase,
		Skill:     params.Skill,
		Status:    string(enums.StepStatusRunning),
		Model:     params.Model,
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
				step.Status = string(enums.StepStatusFailed)
				step.ResultStatus = constants.StepResultFailed
				step.CompletedAt = &completedAt
				step.DurationMs = int(completedAt.Sub(startedAt).Milliseconds())
				step.OutputSummary = yolostrings.Truncate(combinedOutput.String(), 500)
				result.Step = step
				result.Status = constants.StepResultFailed
				result.Error = fmt.Sprintf("command failed: %s: %s", cmd, cmdOut)
				result.Output = combinedOutput.String()
				return result, nil
			}
		}
		completedAt := time.Now()
		step.Status = string(enums.StepStatusCompleted)
		step.ResultStatus = constants.StepResultDone
		step.CompletedAt = &completedAt
		step.DurationMs = int(completedAt.Sub(startedAt).Milliseconds())
		step.OutputSummary = yolostrings.Truncate(combinedOutput.String(), 500)
		result.Step = step
		result.Status = constants.StepResultDone
		result.Output = combinedOutput.String()
		return result, nil
	}

	// Agent step: run Claude.
	claudeResult, err := s.Claude.Run(ctx, params.Config, params.Prompt)

	completedAt := time.Now()
	step.DurationMs = int(completedAt.Sub(startedAt).Milliseconds())
	step.CompletedAt = &completedAt

	if err != nil {
		step.Status = string(enums.StepStatusFailed)
		step.ResultStatus = constants.StepResultFailed
		step.OutputSummary = yolostrings.Truncate(err.Error(), 500)
		result.Step = step
		result.Status = constants.StepResultFailed
		result.Error = err.Error()
		return result, nil
	}

	// Populate step metrics from claude result.
	step.SessionID = claudeResult.SessionID
	step.CostUSD = claudeResult.CostUSD
	step.TokensIn = claudeResult.Usage.InputTokens
	step.TokensOut = claudeResult.Usage.OutputTokens
	step.Turns = claudeResult.NumTurns
	result.SessionID = claudeResult.SessionID
	result.Output = claudeResult.Text

	if claudeResult.IsError {
		step.Status = string(enums.StepStatusFailed)
		step.ResultStatus = constants.StepResultFailed
		step.OutputSummary = yolostrings.Truncate(claudeResult.Text, 500)
		result.Step = step
		result.Status = constants.StepResultFailed
		result.Error = claudeResult.Text
		return result, nil
	}

	// Parse structured output for audit and review steps.
	switch params.Phase {
	case string(enums.StepPhaseAudit):
		failed, errMsg := parseAuditOutput(claudeResult.StructuredOutput)
		if failed {
			step.Status = string(enums.StepStatusFailed)
			step.ResultStatus = constants.StepResultFailed
			step.OutputSummary = yolostrings.Truncate(errMsg, 500)
			result.Step = step
			result.Status = constants.StepResultFailed
			result.Error = errMsg
			return result, nil
		}
	case string(enums.StepPhaseReview):
		review, failed, errMsg := parseReviewOutput(claudeResult.StructuredOutput, params.RunID, params.TaskID, claudeResult)
		result.Review = review
		if failed {
			step.Status = string(enums.StepStatusFailed)
			step.ResultStatus = constants.StepResultFailed
			step.OutputSummary = yolostrings.Truncate(errMsg, 500)
			result.Step = step
			result.Status = constants.StepResultFailed
			result.Error = errMsg
			return result, nil
		}
	}

	step.Status = string(enums.StepStatusCompleted)
	step.ResultStatus = constants.StepResultDone
	step.OutputSummary = yolostrings.Truncate(claudeResult.Text, 500)
	result.Step = step
	result.Status = constants.StepResultDone
	return result, nil
}

// executeLintStep runs zero-token static analysis checks on the project code.
func (s *OrchestratorService) executeLintStep(ctx context.Context, task entities.Task, run *entities.Run, project entities.Project, workDir string, runID string) (*StepResult, error) {
	stepID := ulid.Make().String()
	startedAt := time.Now()

	step := entities.Step{
		RunID:     runID,
		Phase:     string(enums.StepPhaseLint),
		Skill:     "factory-lint",
		Status:    string(enums.StepStatusRunning),
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
		step.Status = string(enums.StepStatusFailed)
		step.ResultStatus = constants.StepResultFailed
		step.CompletedAt = &completedAt
		step.DurationMs = int(completedAt.Sub(startedAt).Milliseconds())
		step.OutputSummary = yolostrings.Truncate(err.Error(), 500)
		return &StepResult{Step: step, Status: constants.StepResultFailed, Error: err.Error()}, nil
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
		step.Status = string(enums.StepStatusFailed)
		step.ResultStatus = constants.StepResultFailed
		step.OutputSummary = yolostrings.Truncate(summary+"\n"+errText, 500)
		return &StepResult{Step: step, Status: constants.StepResultFailed, Error: errText, Output: summary}, nil
	}

	step.Status = string(enums.StepStatusCompleted)
	step.ResultStatus = constants.StepResultDone
	step.OutputSummary = summary
	return &StepResult{Step: step, Status: constants.StepResultDone, Output: summary}, nil
}

// findImplOutput extracts the output from the implement step, used as commit summary.
func (s *OrchestratorService) findImplOutput(steps []entities.Step) string {
	for _, st := range steps {
		if st.Phase == string(enums.StepPhaseImplement) {
			return st.OutputSummary
		}
	}
	return ""
}
