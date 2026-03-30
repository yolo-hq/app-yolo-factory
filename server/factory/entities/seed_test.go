package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yolo-hq/yolo/core/seed/fake"
	seedgraph "github.com/yolo-hq/yolo/core/seed/graph"
)

func TestParseFactoryEntities(t *testing.T) {
	repo, err := fake.ParseEntity(Repo{})
	require.NoError(t, err)
	assert.Equal(t, "Repo", repo.Name)
	assert.Equal(t, "repos", repo.TableName)

	task, err := fake.ParseEntity(Task{})
	require.NoError(t, err)
	assert.Equal(t, "Task", task.Name)
	assert.Equal(t, "tasks", task.TableName)
	require.Len(t, task.Relations, 1)
	assert.Equal(t, "Repo", task.Relations[0].Entity)
	assert.Equal(t, 5, task.Relations[0].Multiplier)

	run, err := fake.ParseEntity(Run{})
	require.NoError(t, err)
	assert.Equal(t, "Run", run.Name)
	assert.Equal(t, "runs", run.TableName)
	require.Len(t, run.Relations, 2)

	question, err := fake.ParseEntity(Question{})
	require.NoError(t, err)
	assert.Equal(t, "Question", question.Name)
	assert.Equal(t, "questions", question.TableName)
	require.Len(t, question.Relations, 3)
}

func TestFactoryDependencyGraph(t *testing.T) {
	entities := parseAll(t)

	g, err := seedgraph.Build(entities)
	require.NoError(t, err)

	order, err := g.TopologicalOrder()
	require.NoError(t, err)

	// Repo must come first (no deps), then Task (depends on Repo),
	// then Run (depends on Task + Repo), then Question (depends on all)
	repoIdx := indexOf(order, "Repo")
	taskIdx := indexOf(order, "Task")
	runIdx := indexOf(order, "Run")
	questionIdx := indexOf(order, "Question")

	assert.True(t, repoIdx < taskIdx, "Repo before Task")
	assert.True(t, taskIdx < runIdx, "Task before Run")
	assert.True(t, runIdx < questionIdx || taskIdx < questionIdx, "Run/Task before Question")
}

func TestFactorySeedPlan(t *testing.T) {
	entities := parseAll(t)

	g, err := seedgraph.Build(entities)
	require.NoError(t, err)

	plans, err := g.Plan(100)
	require.NoError(t, err)

	counts := make(map[string]int)
	for _, p := range plans {
		counts[p.EntityName] = p.Count
	}

	assert.Equal(t, 100, counts["Repo"])
	assert.Equal(t, 500, counts["Task"])  // 100 repos * 5x
	assert.Equal(t, 1000, counts["Run"])  // 500 tasks * 2x
	// Question has no multiplier — gets base size
	assert.Equal(t, 100, counts["Question"])

	total := seedgraph.TotalRows(plans)
	assert.Equal(t, 1700, total) // 100 + 500 + 1000 + 100
}

func TestFactoryGenerate(t *testing.T) {
	entities := parseAll(t)
	gen := fake.NewGenerator(fake.SeedFromSize(100))

	// Generate repos first
	repo := findMeta(entities, "Repo")
	parentIDs := make(map[string][]string)

	for i := 0; i < 10; i++ {
		row := gen.Generate(repo, parentIDs)
		assert.NotEmpty(t, row["id"])
		assert.NotEmpty(t, row["name"])
		assert.NotEmpty(t, row["url"])
		assert.Contains(t, []string{"main", "develop", "staging"}, row["target_branch"])
		parentIDs["Repo"] = append(parentIDs["Repo"], row["id"].(string))
	}

	// Generate tasks (should pick repo IDs)
	task := findMeta(entities, "Task")
	for i := 0; i < 10; i++ {
		row := gen.Generate(task, parentIDs)
		assert.NotEmpty(t, row["id"])
		assert.NotEmpty(t, row["title"])
		assert.Contains(t, []string{"auto", "manual", "review"}, row["type"])
		assert.Contains(t, []string{"queued", "running", "completed", "failed", "cancelled"}, row["status"])
		// repo_id is a multiplier rel — not set by Generate
	}
}

func TestFactoryDeterministic(t *testing.T) {
	entities := parseAll(t)
	repo := findMeta(entities, "Repo")

	gen1 := fake.NewGenerator(fake.SeedFromSize(1000))
	gen2 := fake.NewGenerator(fake.SeedFromSize(1000))

	row1 := gen1.Generate(repo, nil)
	row2 := gen2.Generate(repo, nil)

	assert.Equal(t, row1["name"], row2["name"])
	assert.Equal(t, row1["url"], row2["url"])
	assert.Equal(t, row1["target_branch"], row2["target_branch"])
}

func parseAll(t *testing.T) []*fake.EntityMeta {
	t.Helper()
	var entities []*fake.EntityMeta
	for _, e := range []any{Repo{}, Task{}, Run{}, Question{}} {
		meta, err := fake.ParseEntity(e)
		require.NoError(t, err)
		entities = append(entities, meta)
	}
	return entities
}

func findMeta(entities []*fake.EntityMeta, name string) *fake.EntityMeta {
	for _, e := range entities {
		if e.Name == name {
			return e
		}
	}
	return nil
}

func indexOf(s []string, v string) int {
	for i, e := range s {
		if e == v {
			return i
		}
	}
	return -1
}
