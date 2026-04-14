package actions_test

import (
	"context"
	"os"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/uptrace/bun"
	"github.com/yolo-hq/yolo/core/action"
	bunrepo "github.com/yolo-hq/yolo/core/bun"
	"github.com/yolo-hq/yolo/core/registry"
	"github.com/yolo-hq/yolo/yolotest"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

// TestMain registers entities and repo factories once for all integration tests.
func TestMain(m *testing.M) {
	registry.Register(
		entities.Insight{},
		entities.LintResult{},
		entities.PRD{},
		entities.Project{},
		entities.Question{},
		entities.Review{},
		entities.Run{},
		entities.Step{},
		entities.Suggestion{},
		entities.Task{},
	)

	registry.RegisterRepoFactory("Insight", func(db any) any {
		d := db.(*bun.DB)
		return bunrepo.NewWriteRepository[entities.Insight](d)
	})
	registry.RegisterRepoFactory("LintResult", func(db any) any {
		d := db.(*bun.DB)
		return bunrepo.NewWriteRepository[entities.LintResult](d)
	})
	registry.RegisterRepoFactory("PRD", func(db any) any {
		d := db.(*bun.DB)
		return bunrepo.NewWriteRepository[entities.PRD](d)
	})
	registry.RegisterRepoFactory("Project", func(db any) any {
		d := db.(*bun.DB)
		return bunrepo.NewWriteRepository[entities.Project](d)
	})
	registry.RegisterRepoFactory("Question", func(db any) any {
		d := db.(*bun.DB)
		return bunrepo.NewWriteRepository[entities.Question](d)
	})
	registry.RegisterRepoFactory("Review", func(db any) any {
		d := db.(*bun.DB)
		return bunrepo.NewWriteRepository[entities.Review](d)
	})
	registry.RegisterRepoFactory("Run", func(db any) any {
		d := db.(*bun.DB)
		return bunrepo.NewWriteRepository[entities.Run](d)
	})
	registry.RegisterRepoFactory("Step", func(db any) any {
		d := db.(*bun.DB)
		return bunrepo.NewWriteRepository[entities.Step](d)
	})
	registry.RegisterRepoFactory("Suggestion", func(db any) any {
		d := db.(*bun.DB)
		return bunrepo.NewWriteRepository[entities.Suggestion](d)
	})
	registry.RegisterRepoFactory("Task", func(db any) any {
		d := db.(*bun.DB)
		return bunrepo.NewWriteRepository[entities.Task](d)
	})

	os.Exit(m.Run())
}

// newID returns a new ULID string for test entity IDs.
func newID() string { return ulid.Make().String() }

// runAction executes an action using the test DB transaction.
func runAction(t testing.TB, db bun.IDB, act action.Action, opts ...yolotest.ActionOption) yolotest.ActionResult {
	t.Helper()
	return yolotest.RunActionWithDB(t, db, act, opts...)
}

// seedProject inserts a Project into the tx and returns it.
func seedProject(t testing.TB, tx bun.Tx, overrides *entities.Project) *entities.Project {
	t.Helper()
	p := &entities.Project{
		Name:     "test-project-" + newID(),
		Status:   "active",
		RepoURL:  "https://github.com/test/repo",
		LocalPath: "/tmp/test",
	}
	if overrides != nil {
		if overrides.Status != "" {
			p.Status = overrides.Status
		}
		if overrides.Name != "" {
			p.Name = overrides.Name
		}
	}
	p.ID = newID()
	_, err := tx.NewInsert().Model(p).Exec(context.Background())
	if err != nil {
		t.Fatalf("seedProject: %v", err)
	}
	return p
}

// seedPRD inserts a PRD into the tx and returns it.
func seedPRD(t testing.TB, tx bun.Tx, projectID string, overrides *entities.PRD) *entities.PRD {
	t.Helper()
	p := &entities.PRD{
		ProjectID:          projectID,
		Title:              "Test PRD",
		Body:               "PRD body text",
		AcceptanceCriteria: "Must pass all tests",
		Status:             "draft",
	}
	if overrides != nil {
		if overrides.Status != "" {
			p.Status = overrides.Status
		}
	}
	p.ID = newID()
	_, err := tx.NewInsert().Model(p).Exec(context.Background())
	if err != nil {
		t.Fatalf("seedPRD: %v", err)
	}
	return p
}

// seedTask inserts a Task into the tx and returns it.
func seedTask(t testing.TB, tx bun.Tx, projectID, prdID string, overrides *entities.Task) *entities.Task {
	t.Helper()
	task := &entities.Task{
		ProjectID:          projectID,
		PrdID:              prdID,
		Title:              "Test Task",
		Spec:               "Task spec",
		AcceptanceCriteria: "Must pass",
		Branch:             "feature/test",
		Sequence:           1,
		Status:             "queued",
	}
	if overrides != nil {
		if overrides.Status != "" {
			task.Status = overrides.Status
		}
		if overrides.MaxRetries > 0 {
			task.MaxRetries = overrides.MaxRetries
		}
	}
	task.ID = newID()
	_, err := tx.NewInsert().Model(task).Exec(context.Background())
	if err != nil {
		t.Fatalf("seedTask: %v", err)
	}
	return task
}

// seedInsight inserts an Insight into the tx and returns it.
func seedInsight(t testing.TB, tx bun.Tx, projectID string, overrides *entities.Insight) *entities.Insight {
	t.Helper()
	ins := &entities.Insight{
		ProjectID:      projectID,
		Category:       "retry_rate",
		Title:          "Test Insight",
		Body:           "Insight body",
		Recommendation: "Do something",
		Status:         "pending",
	}
	if overrides != nil {
		if overrides.Status != "" {
			ins.Status = overrides.Status
		}
	}
	ins.ID = newID()
	_, err := tx.NewInsert().Model(ins).Exec(context.Background())
	if err != nil {
		t.Fatalf("seedInsight: %v", err)
	}
	return ins
}

// seedSuggestion inserts a Suggestion into the tx and returns it.
func seedSuggestion(t testing.TB, tx bun.Tx, projectID string, overrides *entities.Suggestion) *entities.Suggestion {
	t.Helper()
	s := &entities.Suggestion{
		ProjectID: projectID,
		Source:    "review",
		Category:  "bug_fix",
		Title:     "Test Suggestion",
		Body:      "Suggestion body",
		Priority:  "medium",
		Status:    "pending",
	}
	if overrides != nil {
		if overrides.Status != "" {
			s.Status = overrides.Status
		}
	}
	s.ID = newID()
	_, err := tx.NewInsert().Model(s).Exec(context.Background())
	if err != nil {
		t.Fatalf("seedSuggestion: %v", err)
	}
	return s
}

// seedRun inserts a Run into the tx and returns it.
func seedRun(t testing.TB, tx bun.Tx, taskID string, overrides *entities.Run) *entities.Run {
	t.Helper()
	r := &entities.Run{
		TaskID:    taskID,
		AgentType: "claude-cli",
		Model:     "sonnet",
		Status:    "running",
	}
	if overrides != nil {
		if overrides.Status != "" {
			r.Status = overrides.Status
		}
	}
	r.ID = newID()
	_, err := tx.NewInsert().Model(r).Exec(context.Background())
	if err != nil {
		t.Fatalf("seedRun: %v", err)
	}
	return r
}

// seedQuestion inserts a Question into the tx and returns it.
func seedQuestion(t testing.TB, tx bun.Tx, taskID, runID string, overrides *entities.Question) *entities.Question {
	t.Helper()
	q := &entities.Question{
		TaskID:     taskID,
		RunID:      runID,
		Body:       "What should I do?",
		Confidence: "medium",
		Status:     "open",
	}
	if overrides != nil {
		if overrides.Status != "" {
			q.Status = overrides.Status
		}
	}
	q.ID = newID()
	_, err := tx.NewInsert().Model(q).Exec(context.Background())
	if err != nil {
		t.Fatalf("seedQuestion: %v", err)
	}
	return q
}

// dbTx opens a test transaction from yolotest.DB, returns the bun.Tx.
func dbTx(t testing.TB) bun.Tx {
	t.Helper()
	db := yolotest.DB(t)
	tx, ok := db.(bun.Tx)
	if !ok {
		t.Fatal("yolotest.DB did not return a bun.Tx")
	}
	return tx
}

// assertEntityStatus reads an entity's status column from the tx.
func assertProjectStatus(t testing.TB, tx bun.Tx, id, wantStatus string) {
	t.Helper()
	var got struct {
		Status string `bun:"status"`
	}
	err := tx.NewSelect().TableExpr("factory_projects").
		ColumnExpr("status").
		Where("id = ?", id).
		Scan(context.Background(), &got)
	if err != nil {
		t.Fatalf("assertProjectStatus: %v", err)
	}
	if got.Status != wantStatus {
		t.Errorf("project status = %q, want %q", got.Status, wantStatus)
	}
}

func assertPRDStatus(t testing.TB, tx bun.Tx, id, wantStatus string) {
	t.Helper()
	var got struct {
		Status string `bun:"status"`
	}
	err := tx.NewSelect().TableExpr("factory_prds").
		ColumnExpr("status").
		Where("id = ?", id).
		Scan(context.Background(), &got)
	if err != nil {
		t.Fatalf("assertPRDStatus: %v", err)
	}
	if got.Status != wantStatus {
		t.Errorf("prd status = %q, want %q", got.Status, wantStatus)
	}
}

func assertTaskStatus(t testing.TB, tx bun.Tx, id, wantStatus string) {
	t.Helper()
	var got struct {
		Status string `bun:"status"`
	}
	err := tx.NewSelect().TableExpr("factory_tasks").
		ColumnExpr("status").
		Where("id = ?", id).
		Scan(context.Background(), &got)
	if err != nil {
		t.Fatalf("assertTaskStatus: %v", err)
	}
	if got.Status != wantStatus {
		t.Errorf("task status = %q, want %q", got.Status, wantStatus)
	}
}

func assertInsightStatus(t testing.TB, tx bun.Tx, id, wantStatus string) {
	t.Helper()
	var got struct {
		Status string `bun:"status"`
	}
	err := tx.NewSelect().TableExpr("factory_insights").
		ColumnExpr("status").
		Where("id = ?", id).
		Scan(context.Background(), &got)
	if err != nil {
		t.Fatalf("assertInsightStatus: %v", err)
	}
	if got.Status != wantStatus {
		t.Errorf("insight status = %q, want %q", got.Status, wantStatus)
	}
}

func assertSuggestionStatus(t testing.TB, tx bun.Tx, id, wantStatus string) {
	t.Helper()
	var got struct {
		Status string `bun:"status"`
	}
	err := tx.NewSelect().TableExpr("factory_suggestions").
		ColumnExpr("status").
		Where("id = ?", id).
		Scan(context.Background(), &got)
	if err != nil {
		t.Fatalf("assertSuggestionStatus: %v", err)
	}
	if got.Status != wantStatus {
		t.Errorf("suggestion status = %q, want %q", got.Status, wantStatus)
	}
}

func assertQuestionStatus(t testing.TB, tx bun.Tx, id, wantStatus string) {
	t.Helper()
	var got struct {
		Status string `bun:"status"`
	}
	err := tx.NewSelect().TableExpr("factory_questions").
		ColumnExpr("status").
		Where("id = ?", id).
		Scan(context.Background(), &got)
	if err != nil {
		t.Fatalf("assertQuestionStatus: %v", err)
	}
	if got.Status != wantStatus {
		t.Errorf("question status = %q, want %q", got.Status, wantStatus)
	}
}

