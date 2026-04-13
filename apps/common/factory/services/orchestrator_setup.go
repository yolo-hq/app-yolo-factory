package services

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/oklog/ulid/v2"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/constants"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/events"
	"github.com/yolo-hq/yolo/core/entity"
)

// setup performs steps 0-5: load entities, budget check, git branch, setup commands,
// create run entity.
//
// Returns (env, nil, nil) on success.
// Returns (nil, &earlyOut, nil) when the budget is exceeded (soft failure, caller returns earlyOut with nil error).
// Returns (nil, nil, err) on hard errors.
func (s *OrchestratorService) setup(ctx context.Context, in OrchestratorInput) (*orchEnv, *OrchestratorOutput, error) {
	// 0a. Load task, PRD, project.
	task, err := s.TaskRead.FindOne(ctx, entity.FindOneOptions{ID: in.TaskID})
	if err != nil {
		return nil, nil, fmt.Errorf("load task: %w", err)
	}
	if task == nil {
		return nil, nil, fmt.Errorf("task %s not found", in.TaskID)
	}

	prd, err := s.PRDRead.FindOne(ctx, entity.FindOneOptions{ID: task.PrdID})
	if err != nil {
		return nil, nil, fmt.Errorf("load prd: %w", err)
	}
	if prd == nil {
		return nil, nil, fmt.Errorf("prd %s not found", task.PrdID)
	}

	project, err := s.ProjectRead.FindOne(ctx, entity.FindOneOptions{ID: task.ProjectID})
	if err != nil {
		return nil, nil, fmt.Errorf("load project: %w", err)
	}
	if project == nil {
		return nil, nil, fmt.Errorf("project %s not found", task.ProjectID)
	}

	// 0b. Update task: status -> "running", started_at, increment run_count.
	now := time.Now()
	_, err = s.TaskWrite.Update(ctx).
		WhereID(task.ID).
		Set(fields.Task.Status.Name(), string(enums.TaskStatusRunning)).
		Set(fields.Task.StartedAt.Name(), now).
		Set(fields.Task.RunCount.Name(), task.RunCount+1).
		Exec(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("update task to running: %w", err)
	}

	inTask := *task
	inPRD := *prd
	inProject := *project

	// 0c. Budget enforcement — check before any work.
	if err := checkBudget(inProject); err != nil {
		events.BudgetExceeded.Emit(ctx, events.BudgetExceededPayload{
			ProjectID:  inProject.ID,
			Spent:      inProject.SpentThisMonthUSD,
			Limit:      inProject.BudgetMonthlyUSD,
			Percentage: 100,
		})
		early := OrchestratorOutput{
			Status:  string(enums.RunStatusFailed),
			Summary: err.Error(),
		}
		return nil, &early, nil
	}

	// 0b. Budget warning — log if approaching limit.
	if inProject.BudgetMonthlyUSD > 0 && inProject.BudgetWarningAt > 0 {
		ratio := inProject.SpentThisMonthUSD / inProject.BudgetMonthlyUSD
		if ratio >= inProject.BudgetWarningAt {
			events.BudgetWarning.Emit(ctx, events.BudgetWarningPayload{
				ProjectID:  inProject.ID,
				Spent:      inProject.SpentThisMonthUSD,
				Limit:      inProject.BudgetMonthlyUSD,
				Percentage: ratio * 100,
			})
			slog.Warn("budget warning",
				"project", inProject.Name,
				"spent", inProject.SpentThisMonthUSD,
				"limit", inProject.BudgetMonthlyUSD,
				"pct", fmt.Sprintf("%.0f%%", ratio*100),
			)
		}
	}

	// 0c. Emit task started event.
	events.TaskStarted.Emit(ctx, inTask.ID)

	// 1. Determine working directory.
	wDir := workingDir(inProject, inTask.ID)

	// 2. Determine model (with escalation support).
	model := determineModel(inTask, inProject)

	// 3. Read CLAUDE.md and wiki from working dir.
	claudeMD := readCLAUDEMD(inProject.LocalPath)
	wikiContent := readProjectWiki(inProject.LocalPath)
	// Append wiki to agent context so all phases benefit from accumulated knowledge.
	if wikiContent != "" {
		claudeMD = claudeMD + "\n\n---\n\n" + wikiContent
	}

	// 4. Git setup: pull latest, create branch.
	branchName := ""
	if inProject.UseWorktrees {
		wtPath := filepath.Join(inProject.LocalPath, ".worktrees", "task-"+inTask.ID)
		gitOut, err := s.Git.Execute(ctx, GitInput{
			Operation: "worktree_add",
			RepoPath:  inProject.LocalPath,
			TaskID:    inTask.ID,
			Path:      wtPath,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("create worktree: %w", err)
		}
		branchName = gitOut.BranchName
		wDir = wtPath
	} else {
		if _, err := s.Git.Execute(ctx, GitInput{
			Operation: "pull",
			RepoPath:  inProject.LocalPath,
			Branch:    inTask.Branch,
		}); err != nil {
			return nil, nil, fmt.Errorf("git pull: %w", err)
		}
		gitOut, err := s.Git.Execute(ctx, GitInput{
			Operation: "branch",
			RepoPath:  inProject.LocalPath,
			TaskID:    inTask.ID,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("create branch: %w", err)
		}
		branchName = gitOut.BranchName
	}

	// 4b. Run setup commands if defined.
	setupCommands := parseTestCommands(inProject.SetupCommands)
	if inProject.SetupCommands != "" && inProject.SetupCommands != "[]" {
		for _, cmd := range setupCommands {
			if _, err := runShellCommand(ctx, wDir, cmd); err != nil {
				return nil, nil, fmt.Errorf("setup command %q: %w", cmd, err)
			}
		}
	}

	// 5. Create Run entity struct.
	runID := ulid.Make().String()
	now = time.Now()
	run := entities.Run{
		TaskID:      inTask.ID,
		AgentType:   constants.AgentImplementer,
		Status:      string(enums.RunStatusRunning),
		Model:       model,
		BranchName:  branchName,
		StartedAt:   now,
		SessionName: fmt.Sprintf("factory:task-%s", inTask.ID),
	}
	run.ID = runID

	// Track escalation.
	if t := inTask; t.RunCount >= inProject.EscalationAfterRetries && inProject.EscalationModel != "" && t.Model == "" {
		run.EscalatedModel = model
	}

	return &orchEnv{
		task:        inTask,
		prd:         inPRD,
		project:     inProject,
		workDir:     wDir,
		branchName:  branchName,
		model:       model,
		claudeMD:    claudeMD,
		wikiContent: wikiContent,
		run:         run,
	}, nil, nil
}

// cleanupOnError marks the task as failed when Execute returns a hard error.
func (e *orchEnv) cleanupOnError(s *OrchestratorService, ctx context.Context, retErr *error) {
	if *retErr != nil {
		if _, uerr := s.TaskWrite.Update(ctx).
			WhereID(e.task.ID).
			Set(fields.Task.Status.Name(), string(enums.TaskStatusFailed)).
			Exec(ctx); uerr != nil {
			slog.Error("failed to mark task as failed after hard error", "task_id", e.task.ID, "error", uerr)
		}
	}
}

// gitCleanup removes the worktree (if used) and deletes the feature branch.
func (s *OrchestratorService) gitCleanup(ctx context.Context, env *orchEnv) {
	if env.project.UseWorktrees {
		wtPath := filepath.Join(env.project.LocalPath, ".worktrees", "task-"+env.task.ID)
		if _, err := s.Git.Execute(ctx, GitInput{
			Operation: "worktree_remove",
			RepoPath:  env.project.LocalPath,
			Path:      wtPath,
		}); err != nil {
			slog.Warn("git worktree cleanup failed", "task_id", env.task.ID, "error", err)
		}
	}
	if _, err := s.Git.Execute(ctx, GitInput{
		Operation: "delete_branch",
		RepoPath:  env.project.LocalPath,
		TaskID:    env.task.ID,
	}); err != nil {
		slog.Warn("git branch cleanup failed", "task_id", env.task.ID, "error", err)
	}
}
