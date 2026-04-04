package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// initTestRepo creates a git repo in a temp dir with an initial commit.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	svc := &GitService{}
	ctx := context.Background()

	_, err := svc.run(ctx, dir, "git", "init")
	require.NoError(t, err)

	_, err = svc.run(ctx, dir, "git", "config", "user.email", "test@test.com")
	require.NoError(t, err)
	_, err = svc.run(ctx, dir, "git", "config", "user.name", "Test")
	require.NoError(t, err)

	// Create initial commit.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("init"), 0644))
	_, err = svc.run(ctx, dir, "git", "add", "-A")
	require.NoError(t, err)
	_, err = svc.run(ctx, dir, "git", "commit", "-m", "init")
	require.NoError(t, err)

	return dir
}

func TestGitService_BranchAndMerge(t *testing.T) {
	dir := initTestRepo(t)
	svc := &GitService{}
	ctx := context.Background()

	// Create task branch.
	out, err := svc.Execute(ctx, GitInput{Operation: "branch", RepoPath: dir, TaskID: "42"})
	require.NoError(t, err)
	assert.Equal(t, "task-42", out.BranchName)

	// Commit a file on the branch.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "feature.go"), []byte("package x"), 0644))
	out, err = svc.Execute(ctx, GitInput{Operation: "commit", RepoPath: dir, Message: "add feature"})
	require.NoError(t, err)
	assert.NotEmpty(t, out.CommitHash)

	// Merge back to main.
	out, err = svc.Execute(ctx, GitInput{Operation: "merge", RepoPath: dir, Branch: "main", TaskID: "42"})
	require.NoError(t, err)

	// Verify feature.go exists on main.
	_, err = os.Stat(filepath.Join(dir, "feature.go"))
	assert.NoError(t, err)
}

func TestGitService_DiffFiles(t *testing.T) {
	dir := initTestRepo(t)
	svc := &GitService{}
	ctx := context.Background()

	// Modify a tracked file.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("changed"), 0644))

	out, err := svc.Execute(ctx, GitInput{Operation: "diff_files", RepoPath: dir})
	require.NoError(t, err)
	assert.Contains(t, out.FilesChanged, "README.md")
}

func TestGitService_WorktreeAddRemove(t *testing.T) {
	dir := initTestRepo(t)
	svc := &GitService{}
	ctx := context.Background()

	wtPath := filepath.Join(t.TempDir(), "wt-99")

	out, err := svc.Execute(ctx, GitInput{
		Operation: "worktree_add",
		RepoPath:  dir,
		TaskID:    "99",
		Path:      wtPath,
	})
	require.NoError(t, err)
	assert.Equal(t, "task-99", out.BranchName)
	assert.Equal(t, wtPath, out.WorktreePath)

	// Worktree dir should exist.
	_, err = os.Stat(wtPath)
	require.NoError(t, err)

	// Remove worktree.
	_, err = svc.Execute(ctx, GitInput{
		Operation: "worktree_remove",
		RepoPath:  dir,
		Path:      wtPath,
	})
	require.NoError(t, err)
}

func TestGitService_DeleteBranch(t *testing.T) {
	dir := initTestRepo(t)
	svc := &GitService{}
	ctx := context.Background()

	// Create then switch away from the branch.
	_, err := svc.Execute(ctx, GitInput{Operation: "branch", RepoPath: dir, TaskID: "7"})
	require.NoError(t, err)
	_, err = svc.Execute(ctx, GitInput{Operation: "checkout", RepoPath: dir, Branch: "main"})
	require.NoError(t, err)

	out, err := svc.Execute(ctx, GitInput{Operation: "delete_branch", RepoPath: dir, TaskID: "7"})
	require.NoError(t, err)
	assert.Equal(t, "task-7", out.BranchName)
}

func TestGitService_UnknownOperation(t *testing.T) {
	svc := &GitService{}
	_, err := svc.Execute(context.Background(), GitInput{Operation: "nope"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown git operation")
}
