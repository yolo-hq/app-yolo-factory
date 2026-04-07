package services

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/yolo-hq/yolo/core/service"
)

// GitService executes git operations for task workflows.
type GitService struct {
	service.Base
}

// GitInput describes a git operation to perform.
type GitInput struct {
	Operation string // pull, checkout, branch, merge, push, commit, worktree_add, worktree_remove, diff_files, delete_branch
	RepoPath  string // working directory for the git command
	Branch    string // target branch (for pull, checkout, merge, push)
	TaskID    string // used to derive branch names (task-{taskID})
	Message   string // commit message (for commit)
	Path      string // worktree path (for worktree_add, worktree_remove)
}

// GitOutput holds the result of a git operation.
type GitOutput struct {
	CommitHash   string
	FilesChanged []string
	BranchName   string
	WorktreePath string
	RawOutput    string
}

// Execute dispatches the git operation based on the Operation field.
func (s *GitService) Execute(ctx context.Context, in GitInput) (GitOutput, error) {
	switch in.Operation {
	case "pull":
		return s.run(ctx, in.RepoPath, "git", "pull", "origin", in.Branch)
	case "checkout":
		return s.run(ctx, in.RepoPath, "git", "checkout", in.Branch)
	case "branch":
		branchName := "task-" + in.TaskID
		out, err := s.run(ctx, in.RepoPath, "git", "checkout", "-b", branchName)
		out.BranchName = branchName
		return out, err
	case "merge":
		branchName := "task-" + in.TaskID
		if _, err := s.run(ctx, in.RepoPath, "git", "checkout", in.Branch); err != nil {
			return GitOutput{}, fmt.Errorf("checkout %s: %w", in.Branch, err)
		}
		out, err := s.run(ctx, in.RepoPath, "git", "merge", branchName)
		out.BranchName = branchName
		return out, err
	case "push":
		return s.run(ctx, in.RepoPath, "git", "push", "origin", in.Branch)
	case "commit":
		if _, err := s.run(ctx, in.RepoPath, "git", "add", "-A"); err != nil {
			return GitOutput{}, fmt.Errorf("git add: %w", err)
		}
		out, err := s.run(ctx, in.RepoPath, "git", "commit", "-m", in.Message)
		if err != nil {
			return out, err
		}
		// Get the commit hash.
		hashOut, err := s.run(ctx, in.RepoPath, "git", "rev-parse", "HEAD")
		if err == nil {
			out.CommitHash = strings.TrimSpace(hashOut.RawOutput)
		}
		return out, nil
	case "worktree_add":
		branchName := "task-" + in.TaskID
		out, err := s.run(ctx, in.RepoPath, "git", "worktree", "add", in.Path, "-b", branchName)
		out.BranchName = branchName
		out.WorktreePath = in.Path
		return out, err
	case "worktree_remove":
		return s.run(ctx, in.RepoPath, "git", "worktree", "remove", in.Path)
	case "diff_files":
		out, err := s.run(ctx, in.RepoPath, "git", "diff", "--name-only")
		if err != nil {
			return out, err
		}
		raw := strings.TrimSpace(out.RawOutput)
		if raw != "" {
			out.FilesChanged = strings.Split(raw, "\n")
		}
		return out, nil
	case "delete_branch":
		branchName := "task-" + in.TaskID
		out, err := s.run(ctx, in.RepoPath, "git", "branch", "-D", branchName)
		out.BranchName = branchName
		return out, err
	default:
		return GitOutput{}, fmt.Errorf("unknown git operation: %s", in.Operation)
	}
}

func (s *GitService) run(ctx context.Context, dir string, name string, args ...string) (GitOutput, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return GitOutput{RawOutput: stderr.String()}, fmt.Errorf("%s %s: %s: %w", name, strings.Join(args, " "), stderr.String(), err)
	}
	return GitOutput{RawOutput: stdout.String()}, nil
}
