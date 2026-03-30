package entities

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yolo-hq/yolo/core/seed"
	"github.com/yolo-hq/yolo/core/seed/fake"
	seedgraph "github.com/yolo-hq/yolo/core/seed/graph"
)

const defaultTestDB = "postgresql://postgres@localhost:5432/yolo_factory?sslmode=disable"

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = defaultTestDB
	}
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	return db
}

func factoryEntities(t *testing.T) []*fake.EntityMeta {
	t.Helper()
	var entities []*fake.EntityMeta
	for _, e := range []any{Repo{}, Task{}, Run{}, Question{}} {
		meta, err := fake.ParseEntity(e)
		require.NoError(t, err)
		entities = append(entities, meta)
	}
	return entities
}

func TestSeedDemo_Integration(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	ctx := context.Background()

	entities := factoryEntities(t)
	runner := seed.NewRunner(db, "seeds", entities)

	// Seed with size 100
	result, err := runner.Demo(ctx, 100)
	require.NoError(t, err)

	assert.True(t, result.GeneratedRows > 0, "should generate rows")
	t.Logf("Seeded %d rows in %s", result.TotalRows, result.Duration)

	// Verify counts
	counts := map[string]int{
		"repos": 100, "tasks": 500, "runs": 1000, "questions": 100,
	}
	for table, expected := range counts {
		var actual int
		err := db.QueryRowContext(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, table)).Scan(&actual)
		require.NoError(t, err)
		assert.Equal(t, expected, actual, "table %s", table)
	}

	// Verify FK integrity — all task.repo_id references a valid repo
	var orphanTasks int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tasks t LEFT JOIN repos r ON t.repo_id = r.id WHERE r.id IS NULL`).Scan(&orphanTasks)
	require.NoError(t, err)
	assert.Equal(t, 0, orphanTasks, "no orphan tasks")

	// Verify FK integrity — all run.task_id references a valid task
	var orphanRuns int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM runs r LEFT JOIN tasks t ON r.task_id = t.id WHERE t.id IS NULL`).Scan(&orphanRuns)
	require.NoError(t, err)
	assert.Equal(t, 0, orphanRuns, "no orphan runs")
}

func TestSeedDemo_Deterministic_Integration(t *testing.T) {
	db := testDB(t)
	defer db.Close()
	ctx := context.Background()

	entities := factoryEntities(t)
	runner := seed.NewRunner(db, "seeds", entities)

	// First seed
	_, err := runner.Demo(ctx, 50)
	require.NoError(t, err)

	var firstName string
	err = db.QueryRowContext(ctx, `SELECT name FROM repos ORDER BY id LIMIT 1`).Scan(&firstName)
	require.NoError(t, err)

	// Second seed (truncates + reseeds with same size → same data)
	_, err = runner.Demo(ctx, 50)
	require.NoError(t, err)

	var secondName string
	err = db.QueryRowContext(ctx, `SELECT name FROM repos ORDER BY id LIMIT 1`).Scan(&secondName)
	require.NoError(t, err)

	assert.Equal(t, firstName, secondName, "deterministic: same size = same data")
}

func TestSeedPlan_Integration(t *testing.T) {
	entities := factoryEntities(t)

	g, err := seedgraph.Build(entities)
	require.NoError(t, err)

	plans, err := g.Plan(1000)
	require.NoError(t, err)

	t.Log("Seed plan for --size 1k:")
	t.Log(seedgraph.FormatPlan(plans))

	total := seedgraph.TotalRows(plans)
	// 1000 repos + 5000 tasks + 10000 runs + 1000 questions = 17000
	assert.Equal(t, 17000, total)
}
