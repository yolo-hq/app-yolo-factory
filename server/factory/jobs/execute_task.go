package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/jobs"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

type ExecuteTaskPayload struct {
	RunID  string `json:"run_id"`
	TaskID string `json:"task_id"`
}

type ExecuteTaskJob struct {
	jobs.Base
	RunRead   entity.ReadRepository[entities.Run]
	RunWrite  entity.WriteRepository[entities.Run]
	TaskRead  entity.ReadRepository[entities.Task]
	TaskWrite entity.WriteRepository[entities.Task]
	RepoRead  entity.ReadRepository[entities.Repo]
}

func (j *ExecuteTaskJob) Name() string { return "factory.execute-task" }

func (j *ExecuteTaskJob) Config() jobs.Config {
	return jobs.Config{
		Queue:      "execution",
		MaxRetries: 1,
		Timeout:    30 * time.Minute,
	}
}

func (j *ExecuteTaskJob) Handle(ctx context.Context, payload []byte) error {
	var p ExecuteTaskPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	run, err := j.RunRead.FindOne(ctx, entity.FindOneOptions{ID: p.RunID})
	if err != nil || run == nil {
		return fmt.Errorf("run %s not found", p.RunID)
	}

	task, err := j.TaskRead.FindOne(ctx, entity.FindOneOptions{ID: p.TaskID})
	if err != nil || task == nil {
		return fmt.Errorf("task %s not found", p.TaskID)
	}

	repo, err := j.RepoRead.FindOne(ctx, entity.FindOneOptions{ID: task.RepoID})
	if err != nil || repo == nil {
		return fmt.Errorf("repo %s not found", task.RepoID)
	}

	workDir := repo.LocalPath
	if workDir == "" {
		return fmt.Errorf("repo %s has no local path", repo.Name)
	}

	// Read CLAUDE.md
	claudeMD := ""
	if data, err := os.ReadFile(filepath.Join(workDir, "CLAUDE.md")); err == nil {
		claudeMD = string(data)
	}

	// Build prompt
	// TODO(ralph): feedback loops from repo config (#9)
	prompt := ComposeWorkerPrompt(claudeMD, task.Body, nil)

	// Git setup
	branch := fmt.Sprintf("task-%s", task.ID)
	runGit(workDir, "checkout", repo.TargetBranch)
	runGit(workDir, "pull", "origin", repo.TargetBranch)
	runGit(workDir, "branch", "-D", branch)
	runGit(workDir, "checkout", "-b", branch)

	// Spawn agent
	start := time.Now()
	cmd := exec.CommandContext(ctx, "claude", "--print",
		"--dangerously-skip-permissions",
		"--no-session-persistence",
		"--model", run.Model,
		"-p", prompt)
	cmd.Dir = workDir

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	agentErr := cmd.Run()
	duration := int(time.Since(start).Seconds())
	cost := parseCost(stdout.String())

	now := time.Now()
	status := "complete"
	errMsg := ""
	if agentErr != nil {
		status = "failed"
		errMsg = agentErr.Error()
	}

	commitHash, _ := runGit(workDir, "rev-parse", "HEAD")

	// Update run
	j.RunWrite.Update(ctx).
		Where(entity.FilterCondition{Field: "id", Operator: entity.OpEq, Value: run.ID}).
		Set("status", status).
		Set("cost", cost).
		Set("duration", duration).
		Set("error", errMsg).
		Set("commit_hash", strings.TrimSpace(commitHash)).
		Set("completed_at", &now).
		Exec(ctx)

	// Update task
	taskStatus := status
	if taskStatus == "complete" {
		taskStatus = "done"
	}
	j.TaskWrite.Update(ctx).
		Where(entity.FilterCondition{Field: "id", Operator: entity.OpEq, Value: task.ID}).
		Set("status", taskStatus).
		Set("cost", task.Cost+cost).
		Exec(ctx)

	// Merge if success
	if status == "complete" {
		runGit(workDir, "checkout", repo.TargetBranch)
		runGit(workDir, "merge", branch, "--no-edit")
		runGit(workDir, "push", "origin", repo.TargetBranch)
		runGit(workDir, "branch", "-D", branch)
	} else {
		runGit(workDir, "checkout", repo.TargetBranch)
	}

	return nil
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func parseCost(output string) float64 {
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "total_cost_usd") {
			var m map[string]any
			if json.Unmarshal([]byte(line), &m) == nil {
				if cost, ok := m["total_cost_usd"].(float64); ok {
					return cost
				}
			}
		}
	}
	return 0
}

// ComposeWorkerPrompt builds the agent prompt from parts.
func ComposeWorkerPrompt(claudeMD, taskBody string, feedbackLoops []string) string {
	var b strings.Builder

	b.WriteString("# FRAMEWORK CONTEXT\n\n")
	b.WriteString("1. All entities MUST embed entity.BaseEntity with TableName() and EntityName()\n")
	b.WriteString("2. All mutations MUST use action pipeline (BaseCreate, BaseUpdate, BaseDelete, TypedInput)\n")
	b.WriteString("3. NEVER use raw http.Handler, http.HandlerFunc, or http.ServeMux\n")
	b.WriteString("4. NEVER use manual SQL — use entity.Repository\n")
	b.WriteString("5. Follow domain structure: server/{domain}/entities/, actions/, queries/\n")
	b.WriteString("6. Framework first: if YOLO doesn't have a pattern you need, flag as question\n\n")

	if claudeMD != "" {
		b.WriteString("# REPO CONTEXT\n\n")
		b.WriteString(claudeMD)
		b.WriteString("\n\n")
	}

	b.WriteString("# TASK\n\n")
	b.WriteString(taskBody)
	b.WriteString("\n\n")

	if len(feedbackLoops) > 0 {
		b.WriteString("# FEEDBACK LOOPS\n\nBefore committing, run ALL:\n")
		for _, loop := range feedbackLoops {
			b.WriteString("```\n" + loop + "\n```\n")
		}
		b.WriteString("Do NOT commit if any fails.\n\n")
	}

	b.WriteString("# COMMIT\n\nPrefix: FACTORY:\n\n")
	b.WriteString("# COMPLETION\n\nIf done, output <promise>COMPLETE</promise>\n")

	return b.String()
}
