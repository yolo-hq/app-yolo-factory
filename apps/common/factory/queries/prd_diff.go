package queries

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/projection"
	"github.com/yolo-hq/yolo/core/query"
	"github.com/yolo-hq/yolo/core/read"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

// prdDiffInfo holds the PRD fields needed for diffing.
type prdDiffInfo struct {
	projection.For[entities.PRD]

	ID        string `field:"id"`
	Title     string `field:"title"`
	ProjectID string `field:"projectId"`
}

// projectPathInfo holds the project local path.
type projectPathInfo struct {
	projection.For[entities.Project]

	LocalPath string `field:"localPath"`
}

// taskCommitRow holds task commit data for diff computation.
type taskCommitRow struct {
	projection.For[entities.Task]

	CommitHash string `field:"commitHash"`
}

// DiffPRDResponse is the typed response for DiffPRDQuery.
type DiffPRDResponse struct {
	PRDID        string `json:"prdId"`
	PRDTitle     string `json:"prdTitle"`
	Diff         string `json:"diff"`
	FilesChanged int    `json:"filesChanged"`
	TasksDone    int    `json:"tasksDone"`
	Commits      int    `json:"commits"`
}

// PrdDiffQuery computes the combined git diff for a completed PRD.
type PrdDiffQuery struct {
	query.Base
	query.TypedInput[inputs.DiffPRDInput]
	query.Returns[DiffPRDResponse]
}

func (q *PrdDiffQuery) Description() string { return "Combined git diff for a completed PRD" }

func (q *PrdDiffQuery) Execute(ctx context.Context, qctx *query.Context) error {
	input := q.Input(qctx)

	prd, err := read.FindOne[prdDiffInfo](ctx, input.PRDID)
	if err != nil {
		return fmt.Errorf("diff-prd: load prd: %w", err)
	}
	if prd.ID == "" {
		return action.FailWithCode("NOT_FOUND", fmt.Sprintf("PRD %s not found", input.PRDID))
	}

	project, err := read.FindOne[projectPathInfo](ctx, prd.ProjectID)
	if err != nil {
		return fmt.Errorf("diff-prd: load project: %w", err)
	}
	if project.LocalPath == "" {
		return action.Fail(fmt.Sprintf("project for PRD %s has no local path", input.PRDID))
	}

	doneTasks, err := read.FindMany[taskCommitRow](ctx,
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
		return q.Respond(qctx, DiffPRDResponse{
			PRDID:     prd.ID,
			PRDTitle:  prd.Title,
			TasksDone: len(doneTasks),
		})
	}

	// TODO: use GitService via DI when available.
	first, last := commits[0], commits[len(commits)-1]
	diff, filesChanged, err := runPRDGitDiff(ctx, project.LocalPath, first, last)
	if err != nil {
		return action.FailWithCode("GIT_ERROR", fmt.Sprintf("diff-prd: %v", err))
	}

	return q.Respond(qctx, DiffPRDResponse{
		PRDID:        prd.ID,
		PRDTitle:     prd.Title,
		Diff:         diff,
		FilesChanged: filesChanged,
		TasksDone:    len(doneTasks),
		Commits:      len(commits),
	})
}

func runPRDGitDiff(ctx context.Context, repoPath, first, last string) (string, int, error) {
	diffRange := fmt.Sprintf("%s^..%s", first, last)
	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "diff", diffRange)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Fallback: try without ^ (first commit might be root).
		if first == last {
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
