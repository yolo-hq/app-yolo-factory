package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/read"
	"github.com/yolo-hq/yolo/core/service"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"gopkg.in/yaml.v3"
)

// BackupService writes entity state to a git-backed directory.
type BackupService struct {
	service.Base
	StatePath string
}

// BackupInput describes what to back up.
// If Trigger is "snapshot" and EntityType is empty, a full snapshot of all entity types is performed.
type BackupInput struct {
	Trigger    string // "task_change", "prd_change", "project_change", "snapshot", "manual"
	EntityType string // "project", "prd", "task", "question", "suggestion" — empty for full snapshot
	EntityID   string
	EntityData any // the actual entity data to serialize
}

// BackupOutput holds the result of a backup operation.
type BackupOutput struct {
	FilePath   string
	CommitHash string
}

// Execute writes the entity to YAML and commits it.
// If Trigger is "snapshot" and EntityType is empty, performs a full snapshot of all entity types.
func (s *BackupService) Execute(ctx context.Context, in BackupInput) (BackupOutput, error) {
	// Full snapshot mode.
	if in.Trigger == "snapshot" && in.EntityType == "" {
		return s.snapshotAll(ctx)
	}

	// 1. Ensure state repo exists.
	if err := s.ensureRepo(ctx); err != nil {
		return BackupOutput{}, fmt.Errorf("ensure repo: %w", err)
	}

	// 2. Determine file path.
	relPath := entityFilePath(in.EntityType, in.EntityID, in.Trigger)
	fullPath := filepath.Join(s.StatePath, relPath)

	// 3. Ensure directory exists.
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return BackupOutput{}, fmt.Errorf("mkdir: %w", err)
	}

	// 4. Marshal entity to YAML.
	data, err := marshalEntity(in.EntityData)
	if err != nil {
		return BackupOutput{}, fmt.Errorf("marshal: %w", err)
	}

	// 5. Write file.
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return BackupOutput{}, fmt.Errorf("write file: %w", err)
	}

	// 6. Git add + commit.
	commitHash, err := s.gitCommit(ctx, relPath, in.Trigger, in.EntityType, in.EntityID)
	if err != nil {
		return BackupOutput{}, fmt.Errorf("git commit: %w", err)
	}

	// 7. Best-effort push.
	_ = s.gitPush(ctx)

	return BackupOutput{
		FilePath:   fullPath,
		CommitHash: commitHash,
	}, nil
}

// Recover reads all YAML files from state path and returns deserialized data.
func (s *BackupService) Recover(_ context.Context) ([]any, error) {
	var results []any

	dirs := []string{"projects", "prds", "tasks", "questions", "suggestions", "snapshots"}
	for _, dir := range dirs {
		dirPath := filepath.Join(s.StatePath, dir)
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read dir %s: %w", dir, err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if !strings.HasSuffix(entry.Name(), ".yml") {
				continue
			}

			data, err := os.ReadFile(filepath.Join(dirPath, entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("read %s/%s: %w", dir, entry.Name(), err)
			}

			var obj map[string]any
			if err := yaml.Unmarshal(data, &obj); err != nil {
				// Try JSON fallback.
				if err2 := json.Unmarshal(data, &obj); err2 != nil {
					return nil, fmt.Errorf("unmarshal %s/%s: %w", dir, entry.Name(), err)
				}
			}

			results = append(results, obj)
		}
	}

	return results, nil
}

// entityFilePath returns the relative path for a backup file.
func entityFilePath(entityType, entityID, trigger string) string {
	if trigger == "daily_snapshot" {
		return filepath.Join("snapshots", time.Now().Format("2006-01-02")+".yml")
	}

	switch entityType {
	case "project":
		return filepath.Join("projects", entityID+".yml")
	case "prd":
		return filepath.Join("prds", "prd-"+entityID+".yml")
	case "task":
		return filepath.Join("tasks", "task-"+entityID+".yml")
	case "question":
		return filepath.Join("questions", "question-"+entityID+".yml")
	case "suggestion":
		return filepath.Join("suggestions", "suggestion-"+entityID+".yml")
	default:
		return filepath.Join("other", entityID+".yml")
	}
}

// marshalEntity serializes entity data to YAML.
func marshalEntity(data any) ([]byte, error) {
	return yaml.Marshal(data)
}

// ensureRepo initializes a git repo at StatePath if one doesn't exist.
func (s *BackupService) ensureRepo(ctx context.Context) error {
	if err := os.MkdirAll(s.StatePath, 0755); err != nil {
		return err
	}

	// Check if already a git repo.
	gitDir := filepath.Join(s.StatePath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return nil
	}

	// git init.
	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = s.StatePath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git init: %s: %w", string(out), err)
	}

	// Create initial commit.
	cmd = exec.CommandContext(ctx, "git", "commit", "--allow-empty", "-m", "init: factory state backup")
	cmd.Dir = s.StatePath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("initial commit: %s: %w", string(out), err)
	}

	return nil
}

// gitCommit stages and commits the file.
func (s *BackupService) gitCommit(ctx context.Context, relPath, trigger, entityType, entityID string) (string, error) {
	// git add.
	cmd := exec.CommandContext(ctx, "git", "add", relPath)
	cmd.Dir = s.StatePath
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git add: %s: %w", string(out), err)
	}

	// git commit.
	msg := fmt.Sprintf("backup: %s %s %s", trigger, entityType, entityID)
	cmd = exec.CommandContext(ctx, "git", "commit", "-m", msg)
	cmd.Dir = s.StatePath
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git commit: %s: %w", string(out), err)
	}

	// Get commit hash.
	cmd = exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = s.StatePath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", nil // non-fatal
	}
	return strings.TrimSpace(string(out)), nil
}

// gitPush pushes to origin (best-effort).
func (s *BackupService) gitPush(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "push")
	cmd.Dir = s.StatePath
	_, err := cmd.CombinedOutput()
	return err
}

// snapshotAll backs up all entities of every type.
func (s *BackupService) snapshotAll(ctx context.Context) (BackupOutput, error) {
	if err := backupEntities[entities.Project](ctx, s, "project"); err != nil {
		return BackupOutput{}, fmt.Errorf("backup projects: %w", err)
	}
	if err := backupEntities[entities.PRD](ctx, s, "prd"); err != nil {
		return BackupOutput{}, fmt.Errorf("backup prds: %w", err)
	}
	if err := backupEntities[entities.Task](ctx, s, "task"); err != nil {
		return BackupOutput{}, fmt.Errorf("backup tasks: %w", err)
	}
	if err := backupEntities[entities.Question](ctx, s, "question"); err != nil {
		return BackupOutput{}, fmt.Errorf("backup questions: %w", err)
	}
	if err := backupEntities[entities.Suggestion](ctx, s, "suggestion"); err != nil {
		return BackupOutput{}, fmt.Errorf("backup suggestions: %w", err)
	}
	return BackupOutput{}, nil
}

// backupEntities loads every entity of type T via read.FindMany and writes
// each one to the backup directory.
func backupEntities[T entity.Entity](ctx context.Context, s *BackupService, entityType string) error {
	rows, err := read.FindMany[T](ctx, read.Limit(10000))
	if err != nil {
		return err
	}
	for _, row := range rows {
		// T is constrained to entity.Entity so GetID is available on a value
		// of type T via a method on a pointer receiver of the concrete type.
		entRow := any(&row).(entity.Entity)
		id := entRow.GetID()
		_, err := s.Execute(ctx, BackupInput{
			Trigger:    "snapshot",
			EntityType: entityType,
			EntityID:   id,
			EntityData: row,
		})
		if err != nil {
			return fmt.Errorf("backup %s %s: %w", entityType, id, err)
		}
	}
	return nil
}

func (s *BackupService) Description() string { return "Create and manage factory state backups" }
