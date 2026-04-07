package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackup_WriteYAML(t *testing.T) {
	dir := t.TempDir()
	svc := &BackupService{StatePath: dir}

	entity := map[string]any{
		"id":   "proj-1",
		"name": "my-project",
	}

	out, err := svc.Execute(context.Background(), BackupInput{
		Trigger:    "project_change",
		EntityType: "project",
		EntityID:   "my-project",
		EntityData: entity,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, out.FilePath)
	assert.NotEmpty(t, out.CommitHash)

	// Verify file exists with correct content.
	data, err := os.ReadFile(out.FilePath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "my-project")
	assert.Contains(t, string(data), "proj-1")
}

func TestBackup_GitInit(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "state")
	svc := &BackupService{StatePath: statePath}

	err := svc.ensureRepo(context.Background())
	require.NoError(t, err)

	// .git directory should exist.
	_, err = os.Stat(filepath.Join(statePath, ".git"))
	assert.NoError(t, err)

	// Calling again should be idempotent.
	err = svc.ensureRepo(context.Background())
	assert.NoError(t, err)
}

func TestBackup_DirectoryStructure(t *testing.T) {
	tests := []struct {
		entityType string
		entityID   string
		trigger    string
		expected   string
	}{
		{"project", "my-app", "project_change", "projects/my-app.yml"},
		{"prd", "123", "prd_change", "prds/prd-123.yml"},
		{"task", "456", "task_change", "tasks/task-456.yml"},
		{"question", "789", "manual", "questions/question-789.yml"},
		{"suggestion", "abc", "manual", "suggestions/suggestion-abc.yml"},
	}

	for _, tt := range tests {
		t.Run(tt.entityType, func(t *testing.T) {
			got := entityFilePath(tt.entityType, tt.entityID, tt.trigger)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestBackup_SnapshotPath(t *testing.T) {
	got := entityFilePath("", "", "daily_snapshot")
	assert.True(t, filepath.Dir(got) == "snapshots")
	assert.Contains(t, got, ".yml")
}

func TestBackup_Recover(t *testing.T) {
	dir := t.TempDir()
	svc := &BackupService{StatePath: dir}

	// Create a project YAML file.
	projDir := filepath.Join(dir, "projects")
	require.NoError(t, os.MkdirAll(projDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(projDir, "test.yml"), []byte("id: proj-1\nname: test\n"), 0644))

	results, err := svc.Recover(context.Background())
	require.NoError(t, err)
	assert.Len(t, results, 1)

	m, ok := results[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "proj-1", m["id"])
	assert.Equal(t, "test", m["name"])
}
