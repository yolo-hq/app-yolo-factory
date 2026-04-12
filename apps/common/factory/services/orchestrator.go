package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/pkg/claude"
	"github.com/yolo-hq/yolo/core/service"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/constants"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/events"
)

// OrchestratorService executes a single task through the full workflow:
// plan -> implement -> test -> audit -> review -> merge.
type OrchestratorService struct {
	service.Base
	Claude      *claude.Client
	Git         *GitService
	Context     *ContextService
	Dependency  *DependencyService
	Linter      *LinterService
	Wiki        *WikiService
	TaskRead    entity.ReadRepository[entities.Task]
	TaskWrite   entity.WriteRepository[entities.Task]
	PRDRead     entity.ReadRepository[entities.PRD]
	ProjectRead entity.ReadRepository[entities.Project]
	RunWrite    entity.WriteRepository[entities.Run]
	StepWrite   entity.WriteRepository[entities.Step]
	ReviewWrite entity.WriteRepository[entities.Review]
}

// OrchestratorInput holds the data needed to execute a task workflow.
type OrchestratorInput struct {
	TaskID string
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
	RunID    string
	TaskID   string
	Phase    string // plan, implement, test, audit, review
	Skill    string
	Model    string
	WorkDir  string
	Prompt   string
	Config   claude.Config
	IsShell  bool     // true for test step (exec.Command, not agent)
	Commands []string // shell commands for test step
}

// StepResult holds the output of a single step.
type StepResult struct {
	Step      entities.Step
	SessionID string
	Output    string
	// Status is one of the StepResult* constants (done, done_with_concerns,
	// needs_context, blocked, failed).  Use the Failed() helper for legacy checks.
	Status   string
	Error    string
	Concerns string // populated when Status == StepResultDoneWithConcerns
	Review   *entities.Review
}

// Failed returns true when the step result is a terminal failure.
func (r *StepResult) Failed() bool { return r.Status == constants.StepResultFailed }

// NeedsEscalation returns true when the run must pause for human input.
func (r *StepResult) NeedsEscalation() bool {
	return r.Status == constants.StepResultNeedsContext || r.Status == constants.StepResultBlocked
}

// orchEnv holds all shared state for a single Execute call.
type orchEnv struct {
	task        entities.Task
	prd         entities.PRD
	project     entities.Project
	workDir     string
	branchName  string
	model       string
	claudeMD    string
	wikiContent string
	run         entities.Run
	steps       []entities.Step
	totalCost   float64
}

// Execute runs the full task workflow and returns the result.
func (s *OrchestratorService) Execute(ctx context.Context, in OrchestratorInput) (retOut OrchestratorOutput, retErr error) {
	out := OrchestratorOutput{Status: string(enums.RunStatusFailed)}

	env, earlyOut, err := s.setup(ctx, in)
	if err != nil {
		return out, err
	}
	// setup returns a non-nil earlyOut on budget exceeded (soft failure, no error).
	if earlyOut != nil {
		return *earlyOut, nil
	}
	defer env.cleanupOnError(s, ctx, &retErr)

	results, questions, err := s.runAllSteps(ctx, env)
	if err != nil {
		return out, err
	}

	// Route based on structured status of the last step executed.
	if len(results) > 0 {
		last := results[len(results)-1]
		switch last.Status {
		case constants.StepResultFailed:
			var review *entities.Review
			if last.Review != nil {
				review = last.Review
			}
			return s.buildFailure(ctx, env.task, env.project, env.run, env.steps, review, env.totalCost, last)
		case constants.StepResultBlocked, constants.StepResultNeedsContext:
			var review *entities.Review
			if last.Review != nil {
				review = last.Review
			}
			return s.buildEscalation(ctx, env.task, env.project, env.run, env.steps, review, env.totalCost, last, questions)
		}
	}

	// Collect review from results.
	var review *entities.Review
	for _, r := range results {
		if r.Review != nil {
			review = r.Review
		}
	}

	return s.commitAndPersist(ctx, env, questions, review)
}

// commitAndPersist handles steps 11-18: commit, merge, push, cleanup, events, persist.
func (s *OrchestratorService) commitAndPersist(ctx context.Context, env *orchEnv, questions []entities.Question, review *entities.Review) (OrchestratorOutput, error) {
	out := OrchestratorOutput{Status: string(enums.RunStatusFailed)}

	// 11. Capture files changed BEFORE committing.
	filesOut, _ := s.Git.Execute(ctx, GitInput{
		Operation: "diff_files",
		RepoPath:  env.workDir,
	})
	filesChanged := filesOut.FilesChanged

	// 12. Git commit.
	implOutput := s.findImplOutput(env.steps)
	summary := truncate500(implOutput)
	commitOut, err := s.Git.Execute(ctx, GitInput{
		Operation: "commit",
		RepoPath:  env.workDir,
		Message:   fmt.Sprintf("feat: %s\n\nTask: %s", env.task.Title, env.task.ID),
	})
	if err != nil {
		return out, fmt.Errorf("git commit: %w", err)
	}
	commitHash := commitOut.CommitHash

	// 13. Merge to target branch + push (if auto_merge).
	if env.project.AutoMerge {
		if _, err := s.Git.Execute(ctx, GitInput{
			Operation: "merge",
			RepoPath:  env.project.LocalPath,
			Branch:    env.task.Branch,
			TaskID:    env.task.ID,
		}); err != nil {
			return out, fmt.Errorf("git merge: %w", err)
		}

		if _, err := s.Git.Execute(ctx, GitInput{
			Operation: "push",
			RepoPath:  env.project.LocalPath,
			Branch:    env.task.Branch,
		}); err != nil {
			return out, fmt.Errorf("git push: %w", err)
		}
	}

	// 14. Git cleanup.
	s.gitCleanup(ctx, env)

	// 16. Emit task completed event.
	events.TaskCompleted.Emit(ctx, env.task.ID)

	// 17. Build final output.
	completedAt := time.Now()
	env.run.Status = string(enums.RunStatusCompleted)
	env.run.CostUSD = env.totalCost
	env.run.CommitHash = commitHash
	env.run.FilesChanged = toJSON(filesChanged)
	env.run.CompletedAt = &completedAt
	env.run.Result = summary

	// 18. Persist run, steps, review.
	if _, err := s.RunWrite.Insert(ctx, &env.run); err != nil {
		return out, fmt.Errorf("insert run: %w", err)
	}
	for i := range env.steps {
		if _, err := s.StepWrite.Insert(ctx, &env.steps[i]); err != nil {
			return out, fmt.Errorf("insert step %s: %w", env.steps[i].Phase, err)
		}
	}
	if review != nil {
		if _, err := s.ReviewWrite.Insert(ctx, review); err != nil {
			return out, fmt.Errorf("insert review: %w", err)
		}
	}

	// 19. Update project wiki (non-fatal).
	if s.Wiki != nil {
		reviewText := ""
		if review != nil {
			reviewText = review.Suggestions
		}
		if wikiErr := s.Wiki.Update(ctx, WikiInput{
			ProjectPath:  env.project.LocalPath,
			TaskID:       env.task.ID,
			TaskTitle:    env.task.Title,
			TaskOutput:   implOutput,
			ReviewOutput: reviewText,
		}); wikiErr != nil {
			slog.Warn("wiki update failed (non-fatal)", "task_id", env.task.ID, "error", wikiErr)
		}
	}

	return OrchestratorOutput{
		Run:          env.run,
		Steps:        env.steps,
		Review:       review,
		Questions:    questions,
		Status:       string(enums.RunStatusCompleted),
		CostUSD:      env.totalCost,
		CommitHash:   commitHash,
		FilesChanged: filesChanged,
		Summary:      summary,
	}, nil
}

func (s *OrchestratorService) Description() string {
	return "Execute full implementation workflow for a task"
}
