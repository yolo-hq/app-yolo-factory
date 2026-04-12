package services

import (
	"context"
	"log/slog"
	"time"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
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
