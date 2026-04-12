package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/oklog/ulid/v2"
	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/constants"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/events"
	"github.com/yolo-hq/yolo/core/service"
)

// buildFailure constructs an OrchestratorOutput for early-return (failure) cases.
// It also emits TaskFailed event and pushes the branch if push_failed_branches is set.
func (s *OrchestratorService) buildFailure(ctx context.Context, task entities.Task, project entities.Project, run entities.Run, steps []entities.Step, review *entities.Review, totalCost float64, lastResult *StepResult) (OrchestratorOutput, error) {
	completedAt := time.Now()
	run.Status = string(enums.RunStatusFailed)
	run.CostUSD = totalCost
	run.Error = lastResult.Error
	run.CompletedAt = &completedAt

	service.EmitEvent(ctx, service.PendingEvent{
		EntityType: "Task",
		EntityID:   run.TaskID,
		Name:       events.TaskFailedName,
		Data: events.TaskPayload{
			TaskID:      run.TaskID,
			Title:       task.Title,
			ProjectName: project.Name,
			CostUSD:     totalCost,
			Error:       lastResult.Error,
		},
	})

	if project.PushFailedBranches && run.BranchName != "" {
		if _, err := s.Git.Execute(ctx, GitInput{
			Operation: "push",
			RepoPath:  workingDir(project, task.ID),
			Branch:    run.BranchName,
		}); err != nil {
			slog.Warn("push failed branch failed", "task_id", task.ID, "branch", run.BranchName, "error", err)
		}
	}

	// Persist run and steps on failure.
	if _, err := s.RunWrite.Insert(ctx, &run); err != nil {
		slog.Error("insert failed run", "error", err)
	}
	for i := range steps {
		if _, err := s.StepWrite.Insert(ctx, &steps[i]); err != nil {
			slog.Error("insert step on failure", "phase", steps[i].Phase, "error", err)
		}
	}
	if review != nil {
		if _, err := s.ReviewWrite.Insert(ctx, review); err != nil {
			slog.Error("insert review on failure", "error", err)
		}
	}

	return OrchestratorOutput{
		Run:     run,
		Steps:   steps,
		Review:  review,
		Status:  string(enums.RunStatusFailed),
		CostUSD: totalCost,
		Summary: lastResult.Error,
	}, nil
}

// buildEscalation handles needs_context and blocked step results.
//
//   - needs_context → triggers the existing question escalation flow by appending
//     a Question detected from the step output.
//   - blocked → creates a new Question with "blocked" context so a human can unblock.
//
// In both cases the run is persisted as "blocked" (not "failed") and a TaskBlocked
// event is emitted so external systems can page/notify.
func (s *OrchestratorService) buildEscalation(
	ctx context.Context,
	task entities.Task,
	project entities.Project,
	run entities.Run,
	steps []entities.Step,
	review *entities.Review,
	totalCost float64,
	lastResult *StepResult,
	existingQuestions []entities.Question,
) (OrchestratorOutput, error) {
	completedAt := time.Now()
	run.Status = constants.RunStatusBlocked
	run.CostUSD = totalCost
	run.Error = lastResult.Error
	run.CompletedAt = &completedAt

	// Build question for the escalation.
	var questions []entities.Question
	questions = append(questions, existingQuestions...)

	switch lastResult.Status {
	case constants.StepResultBlocked:
		// Create a blocking question so a human can intervene.
		q := entities.Question{
			TaskID:     task.ID,
			RunID:      run.ID,
			Body:       fmt.Sprintf("Task is blocked at %s step: %s", lastResult.Step.Phase, lastResult.Error),
			Context:    "blocked",
			Confidence: string(enums.QuestionConfidenceMedium),
			Status:     string(enums.QuestionStatusOpen),
		}
		q.ID = ulid.Make().String()
		questions = append(questions, q)
	case constants.StepResultNeedsContext:
		// Question should already be in existingQuestions (detected from output),
		// but if not, synthesise one.
		if len(existingQuestions) == 0 {
			q := entities.Question{
				TaskID:     task.ID,
				RunID:      run.ID,
				Body:       lastResult.Error,
				Context:    "needs_context",
				Confidence: string(enums.QuestionConfidenceMedium),
				Status:     string(enums.QuestionStatusOpen),
			}
			q.ID = ulid.Make().String()
			questions = append(questions, q)
		}
	}

	service.EmitEvent(ctx, service.PendingEvent{
		EntityType: "Task",
		EntityID:   run.TaskID,
		Name:       events.TaskBlockedName,
		Data: events.TaskPayload{
			TaskID:      run.TaskID,
			Title:       task.Title,
			ProjectName: project.Name,
			CostUSD:     totalCost,
			Error:       lastResult.Error,
		},
	})

	// Persist run, steps, and any review produced so far.
	if _, err := s.RunWrite.Insert(ctx, &run); err != nil {
		slog.Error("insert blocked run", "error", err)
	}
	for i := range steps {
		if _, err := s.StepWrite.Insert(ctx, &steps[i]); err != nil {
			slog.Error("insert step on escalation", "phase", steps[i].Phase, "error", err)
		}
	}
	if review != nil {
		if _, err := s.ReviewWrite.Insert(ctx, review); err != nil {
			slog.Error("insert review on escalation", "error", err)
		}
	}

	return OrchestratorOutput{
		Run:       run,
		Steps:     steps,
		Review:    review,
		Questions: questions,
		Status:    constants.RunStatusBlocked,
		CostUSD:   totalCost,
		Summary:   lastResult.Error,
	}, nil
}
