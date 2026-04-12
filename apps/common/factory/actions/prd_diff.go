package actions

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/projection"
	"github.com/yolo-hq/yolo/core/read"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

// PRDDiffInfo holds the PRD fields needed for diffing.
type PRDDiffInfo struct {
	projection.For[entities.PRD]

	ID        string `field:"id"`
	Title     string `field:"title"`
	ProjectID string `field:"projectId"`
}

// ProjectPathInfo holds the project local path.
type ProjectPathInfo struct {
	projection.For[entities.Project]

	LocalPath string `field:"localPath"`
}

// TaskCommitRow holds task commit data for diff computation.
type TaskCommitRow struct {
	projection.For[entities.Task]

	CommitHash string `field:"commitHash"`
}

// DiffPRDResponse is the typed response for DiffPRDAction.
type DiffPRDResponse struct {
	PRDID        string `json:"prdId"`
	PRDTitle     string `json:"prdTitle"`
	Diff         string `json:"diff"`
	FilesChanged int    `json:"filesChanged"`
	TasksDone    int    `json:"tasksDone"`
	Commits      int    `json:"commits"`
}

// DiffPRDAction computes the combined git diff for a completed PRD.
type DiffPRDAction struct {
	action.SkipAllPolicies
	action.TypedInput[inputs.DiffPRDInput]
	action.TypedResponse[DiffPRDResponse]
}

func (a *DiffPRDAction) Description() string { return "Show combined git diff for a completed PRD" }

func (a *DiffPRDAction) Execute(ctx context.Context, actx *action.Context) error {
	input := a.Input(actx)

	prd, err := read.FindOne[PRDDiffInfo](ctx, input.PRDID)
	if err != nil {
		return fmt.Errorf("diff-prd: load prd: %w", err)
	}
	if prd.ID == "" {
		return fmt.Errorf("PRD %s not found", input.PRDID)
	}

	project, err := read.FindOne[ProjectPathInfo](ctx, prd.ProjectID)
	if err != nil {
		return fmt.Errorf("diff-prd: load project: %w", err)
	}
	if project.LocalPath == "" {
		return fmt.Errorf("project for PRD %s has no local path", input.PRDID)
	}

	doneTasks, err := read.FindMany[TaskCommitRow](ctx,
		read.Eq("prd_id", prd.ID),
		read.Eq("status", string(enums.TaskStatusDone)),
		read.OrderBy("sequence", read.Asc),
	)
	if err != nil {
		return fmt.Errorf("diff-prd: list tasks: %w", err)
	}

	var commits []string
	for _, t := range doneTasks {
		if t.CommitHash != "" {
			commits = append(commits, t.CommitHash)
		}
	}

	if len(commits) == 0 {
		return a.Respond(actx, DiffPRDResponse{
			PRDID:     prd.ID,
			PRDTitle:  prd.Title,
			TasksDone: len(doneTasks),
		})
	}

	first, last := commits[0], commits[len(commits)-1]
	diff, filesChanged, err := runGitDiff(ctx, project.LocalPath, first, last)
	if err != nil {
		return fmt.Errorf("diff-prd: %w", err)
	}

	return a.Respond(actx, DiffPRDResponse{
		PRDID:        prd.ID,
		PRDTitle:     prd.Title,
		Diff:         diff,
		FilesChanged: filesChanged,
		TasksDone:    len(doneTasks),
		Commits:      len(commits),
	})
}

func runGitDiff(ctx context.Context, repoPath, first, last string) (string, int, error) {
	diffRange := fmt.Sprintf("%s^..%s", first, last)
	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "diff", diffRange)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Fallback: try without ^ (first commit might be root).
		if len(first) > 0 && first == last {
			diffRange = fmt.Sprintf("%s^..%s", first, first)
		} else {
			diffRange = fmt.Sprintf("%s..%s", first, last)
		}
		stdout.Reset()
		stderr.Reset()
		cmd2 := exec.CommandContext(ctx, "git", "-C", repoPath, "diff", diffRange)
		cmd2.Stdout = &stdout
		cmd2.Stderr = &stderr
		if err2 := cmd2.Run(); err2 != nil {
			return "", 0, fmt.Errorf("git diff: %s: %w", strings.TrimSpace(stderr.String()), err2)
		}
	}

	diff := stdout.String()
	filesChanged := 0
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "diff --git") {
			filesChanged++
		}
	}

	return diff, filesChanged, nil
}
