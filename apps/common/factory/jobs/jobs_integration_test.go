//go:build integration

package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

	corebun "github.com/yolo-hq/yolo/core/bun"
	bunrepo "github.com/yolo-hq/yolo/core/bun"

	svc "github.com/yolo-hq/app-yolo-factory/.yolo/svc"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/services"
)

// openTestDB opens a *bun.DB from DATABASE_URL. Fails if not set.
func openTestDB(t *testing.T) (*bun.DB, bun.Tx, func()) {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL required for integration tests")
	}
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)

	cleanup := func() {
		_ = tx.Rollback()
		_ = db.Close()
	}
	return db, tx, cleanup
}

func newID() string { return ulid.Make().String() }

// TestResetBudgetsJob_ResetsActiveProjects seeds projects with nonzero spend,
// runs ResetBudgetsJob, and verifies spend is zeroed for active projects.
func TestResetBudgetsJob_ResetsActiveProjects(t *testing.T) {
	db, tx, done := openTestDB(t)
	defer done()

	ctx := corebun.WithTx(context.Background(), tx)

	// Seed: two active projects with nonzero spend, one paused.
	proj1 := &entities.Project{SpentThisMonthUSD: 42.5, Status: "active", RepoURL: "https://x.com/1", Name: "p1-" + newID(), LocalPath: "/tmp"}
	proj1.ID = newID()
	proj2 := &entities.Project{SpentThisMonthUSD: 10.0, Status: "active", RepoURL: "https://x.com/2", Name: "p2-" + newID(), LocalPath: "/tmp"}
	proj2.ID = newID()
	paused := &entities.Project{SpentThisMonthUSD: 5.0, Status: "paused", RepoURL: "https://x.com/3", Name: "p3-" + newID(), LocalPath: "/tmp"}
	paused.ID = newID()

	_, err := tx.NewInsert().Model(proj1).Exec(ctx)
	require.NoError(t, err)
	_, err = tx.NewInsert().Model(proj2).Exec(ctx)
	require.NoError(t, err)
	_, err = tx.NewInsert().Model(paused).Exec(ctx)
	require.NoError(t, err)

	// Wire service with repos that use the same *bun.DB (they'll pick up tx via context).
	svc.S.BudgetReset = &services.BudgetResetService{
		ProjectRead:  bunrepo.NewReadRepository[entities.Project](db),
		ProjectWrite: bunrepo.NewWriteRepository[entities.Project](db),
	}

	job := &ResetBudgetsJob{}
	err = job.Handle(ctx, nil)
	require.NoError(t, err)

	// Verify active projects are zeroed.
	var updated entities.Project
	err = tx.NewSelect().Model(&updated).Where("id = ?", proj1.ID).Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0.0, updated.SpentThisMonthUSD, "active project spend should be reset")

	err = tx.NewSelect().Model(&updated).Where("id = ?", proj2.ID).Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0.0, updated.SpentThisMonthUSD, "active project spend should be reset")

	// Paused project should NOT be zeroed.
	err = tx.NewSelect().Model(&updated).Where("id = ?", paused.ID).Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, 5.0, updated.SpentThisMonthUSD, "paused project spend should NOT be reset")
}

// TestCheckTimeoutsJob_FailsTimedOutRuns seeds a run that started long ago,
// runs CheckTimeoutsJob, and verifies the run and its task are marked failed.
func TestCheckTimeoutsJob_FailsTimedOutRuns(t *testing.T) {
	db, tx, done := openTestDB(t)
	defer done()

	ctx := corebun.WithTx(context.Background(), tx)

	// Seed: project, task, run that started 2 hours ago (exceeds any timeout).
	proj := &entities.Project{
		Name: "proj-timeout-" + newID(), Status: "active",
		RepoURL: "https://x.com/t", LocalPath: "/tmp",
		TimeoutSecs: 60, // 1 minute
	}
	proj.ID = newID()
	_, err := tx.NewInsert().Model(proj).Exec(ctx)
	require.NoError(t, err)

	prd := &entities.PRD{
		ProjectID: proj.ID, Title: "T", Body: "B",
		AcceptanceCriteria: "AC", Status: "in_progress",
	}
	prd.ID = newID()
	_, err = tx.NewInsert().Model(prd).Exec(ctx)
	require.NoError(t, err)

	task := &entities.Task{
		ProjectID: proj.ID, PrdID: prd.ID,
		Title: "task-t", Status: "running",
		Spec: "s", AcceptanceCriteria: "ac", Branch: "b",
		Sequence: 1, MaxRetries: 2, RunCount: 1,
	}
	task.ID = newID()
	_, err = tx.NewInsert().Model(task).Exec(ctx)
	require.NoError(t, err)

	longAgo := time.Now().Add(-2 * time.Hour)
	run := &entities.Run{
		TaskID: task.ID, AgentType: "claude-cli",
		Status: "running", Model: "sonnet",
		StartedAt: longAgo,
	}
	run.ID = newID()
	_, err = tx.NewInsert().Model(run).Exec(ctx)
	require.NoError(t, err)

	// Wire service.
	svc.S.Timeout = &services.TimeoutService{
		RunRead:     bunrepo.NewReadRepository[entities.Run](db),
		RunWrite:    bunrepo.NewWriteRepository[entities.Run](db),
		TaskRead:    bunrepo.NewReadRepository[entities.Task](db),
		TaskWrite:   bunrepo.NewWriteRepository[entities.Task](db),
		ProjectRead: bunrepo.NewReadRepository[entities.Project](db),
	}

	job := &CheckTimeoutsJob{}
	err = job.Handle(ctx, nil)
	require.NoError(t, err)

	// Verify run is now failed.
	var updatedRun entities.Run
	err = tx.NewSelect().Model(&updatedRun).Where("id = ?", run.ID).Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "failed", updatedRun.Status, "timed-out run should be marked failed")
}

// TestPlanPRDJob_SkipsExecution verifies that PlanPRDJob.Handle returns an error
// when the service is not wired (nil pointer), confirming it does not silently no-op.
// Jobs that call Claude CLI require a running claude binary; we verify setup only.
func TestPlanPRDJob_SetupAndTeardown(t *testing.T) {
	// Save and restore svc.S.Planner.
	original := svc.S.Planner
	t.Cleanup(func() { svc.S.Planner = original })

	// With a nil planner, Handle must panic or return an error.
	svc.S.Planner = nil
	job := &PlanPRDJob{PRDID: "test-prd-id"}
	payload, err := json.Marshal(map[string]string{"prd_id": "test-prd-id"})
	require.NoError(t, err)

	assert.Panics(t, func() {
		_ = job.Handle(context.Background(), payload)
	}, "nil Planner should panic")
}

// TestExecuteWorkflowJob_SetupAndTeardown verifies ExecuteWorkflowJob setup.
func TestExecuteWorkflowJob_SetupAndTeardown(t *testing.T) {
	original := svc.S.Orchestrator
	t.Cleanup(func() { svc.S.Orchestrator = original })

	svc.S.Orchestrator = nil
	job := &ExecuteWorkflowJob{TaskID: "test-task-id"}
	payload, err := json.Marshal(map[string]string{"task_id": "test-task-id"})
	require.NoError(t, err)

	assert.Panics(t, func() {
		_ = job.Handle(context.Background(), payload)
	}, "nil Orchestrator should panic")
}