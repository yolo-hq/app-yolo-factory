package commands

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/yolo-hq/yolo/core/command"
	"github.com/yolo-hq/yolo/core/entity"

	enums "github.com/yolo-hq/app-yolo-factory/.yolo/enums"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// --- PRDDiff ---

type PRDDiff struct {
	command.Base
}

type PRDDiffInput struct {
	PRDID string `flag:"id" validate:"required" usage:"PRD ID"`
}

func (c *PRDDiff) Name() string        { return "prd:diff" }
func (c *PRDDiff) Description() string { return "Show combined diff for a completed PRD" }
func (c *PRDDiff) Input() any          { return &PRDDiffInput{} }

func (c *PRDDiff) Execute(ctx context.Context, cctx command.Context) error {
	input, _ := cctx.TypedInput.(*PRDDiffInput)

	// Load PRD.
	prdRepo, err := cctx.RepoProvider.Repo("PRD")
	if err != nil {
		return fmt.Errorf("get prd repo: %w", err)
	}
	pr := prdRepo.(entity.ReadRepository[entities.PRD])

	prd, err := pr.FindOne(ctx, entity.FindOneOptions{ID: input.PRDID})
	if err != nil {
		return fmt.Errorf("find prd: %w", err)
	}
	if prd == nil {
		return fmt.Errorf("PRD %s not found", input.PRDID)
	}

	// Load done tasks ordered by sequence.
	taskRepo, err := cctx.RepoProvider.Repo("Task")
	if err != nil {
		return fmt.Errorf("get task repo: %w", err)
	}
	tr := taskRepo.(entity.ReadRepository[entities.Task])

	result, err := tr.FindMany(ctx, entity.FindOptions{
		Filters: []entity.FilterCondition{
			{Field: "prd_id", Operator: entity.OpEq, Value: prd.ID},
			{Field: "status", Operator: entity.OpEq, Value: string(enums.TaskStatusDone)},
		},
		Sort: &entity.SortParams{Field: "sequence", Order: "asc"},
	})
	if err != nil {
		return fmt.Errorf("list tasks: %w", err)
	}

	if len(result.Data) == 0 {
		cctx.Print("No completed tasks found for PRD %s", prd.ID)
		return nil
	}

	// Collect commit hashes.
	var commits []string
	for _, t := range result.Data {
		if t.CommitHash != "" {
			commits = append(commits, t.CommitHash)
		}
	}

	if len(commits) == 0 {
		cctx.Print("No commits found for PRD %s (%d tasks completed but no commit hashes)", prd.ID, len(result.Data))
		return nil
	}

	// Load project for local path.
	projectRepo, err := cctx.RepoProvider.Repo("Project")
	if err != nil {
		return fmt.Errorf("get project repo: %w", err)
	}
	projR := projectRepo.(entity.ReadRepository[entities.Project])

	project, err := projR.FindOne(ctx, entity.FindOneOptions{ID: prd.ProjectID})
	if err != nil {
		return fmt.Errorf("find project: %w", err)
	}
	if project == nil {
		return fmt.Errorf("project %s not found", prd.ProjectID)
	}

	// Run git diff first-commit^..last-commit.
	firstCommit := commits[0]
	lastCommit := commits[len(commits)-1]

	diffRange := fmt.Sprintf("%s^..%s", firstCommit, lastCommit)
	cmd := exec.CommandContext(ctx, "git", "-C", project.LocalPath, "diff", diffRange)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Try without ^ (first commit might be root).
		diffRange = fmt.Sprintf("%s..%s", firstCommit, lastCommit)
		if len(commits) == 1 {
			diffRange = fmt.Sprintf("%s^..%s", firstCommit, firstCommit)
		}
		cmd2 := exec.CommandContext(ctx, "git", "-C", project.LocalPath, "diff", diffRange)
		stdout.Reset()
		stderr.Reset()
		cmd2.Stdout = &stdout
		cmd2.Stderr = &stderr
		if err := cmd2.Run(); err != nil {
			return fmt.Errorf("git diff: %s: %w", strings.TrimSpace(stderr.String()), err)
		}
	}

	diff := stdout.String()
	if diff == "" {
		cctx.Print("No diff output (commits may be empty or identical)")
		return nil
	}

	// Print the diff.
	cctx.Print("%s", diff)

	// Count files changed.
	filesChanged := 0
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "diff --git") {
			filesChanged++
		}
	}

	cctx.Print("---")
	cctx.Print("PRD: %s — %s", prd.ID, prd.Title)
	cctx.Print("%d files changed, %d tasks completed, %d commits", filesChanged, len(result.Data), len(commits))
	return nil
}
