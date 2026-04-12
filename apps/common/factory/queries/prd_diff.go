package queries

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/yolo-hq/yolo/core/projection"
	"github.com/yolo-hq/yolo/core/query"
	"github.com/yolo-hq/yolo/core/read"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
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

// DiffPRDQuery computes the combined git diff for a completed PRD.
type DiffPRDQuery struct {
	query.Base

	PRDID string `arg:"prdId" validate:"required"`
}

func (q *DiffPRDQuery) Execute(ctx context.Context, qctx *query.Context) query.Result {
	prd, err := read.FindOne[PRDDiffInfo](ctx, q.PRDID)
	if err != nil {
		return query.Fail("read_error", fmt.Sprintf("diff-prd: load prd: %v", err))
	}
	if prd.ID == "" {
		return query.Fail("not_found", fmt.Sprintf("PRD %s not found", q.PRDID))
	}

	project, err := read.FindOne[ProjectPathInfo](ctx, prd.ProjectID)
	if err != nil {
		return query.Fail("read_error", fmt.Sprintf("diff-prd: load project: %v", err))
	}
	if project.LocalPath == "" {
		return query.Fail("invalid_state", fmt.Sprintf("project for PRD %s has no local path", q.PRDID))
	}

	doneTasks, err := read.FindMany[TaskCommitRow](ctx,
		read.Eq("prd_id", prd.ID),
		read.Eq("status", string(enums.TaskStatusDone)),
		read.OrderBy("sequence", read.Asc),
	)
	if err != nil {
		return query.Fail("read_error", fmt.Sprintf("diff-prd: list tasks: %v", err))
	}

	var commits []string
	for _, t := range doneTasks {
		if t.CommitHash != "" {
			commits = append(commits, t.CommitHash)
		}
	}

	if len(commits) == 0 {
		return query.OK(query.Extras{
			"prdId":     prd.ID,
			"prdTitle":  prd.Title,
			"diff":      "",
			"tasksDone": len(doneTasks),
			"commits":   0,
		})
	}

	first, last := commits[0], commits[len(commits)-1]
	diff, filesChanged, err := runGitDiff(ctx, project.LocalPath, first, last)
	if err != nil {
		return query.Fail("git_error", fmt.Sprintf("diff-prd: %v", err))
	}

	return query.OK(query.Extras{
		"prdId":        prd.ID,
		"prdTitle":     prd.Title,
		"diff":         diff,
		"filesChanged": filesChanged,
		"tasksDone":    len(doneTasks),
		"commits":      len(commits),
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
