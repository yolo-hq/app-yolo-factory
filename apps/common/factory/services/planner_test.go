package services

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

func TestParseTaskDefs(t *testing.T) {
	raw := json.RawMessage(`{
		"tasks": [
			{
				"title": "Setup database schema",
				"spec": "Create migration for users table",
				"acceptance_criteria": [
					{"id": "AC1", "description": "Migration runs without error"},
					{"id": "AC2", "description": "Table has correct columns"}
				],
				"sequence": 1,
				"depends_on": [],
				"estimated_complexity": "low"
			},
			{
				"title": "Implement user CRUD",
				"spec": "Build actions for user entity",
				"acceptance_criteria": [
					{"id": "AC1", "description": "All CRUD endpoints work"}
				],
				"sequence": 2,
				"depends_on": [1],
				"estimated_complexity": "medium"
			}
		]
	}`)

	defs, err := parseTaskDefs(raw)
	require.NoError(t, err)
	assert.Len(t, defs, 2)

	assert.Equal(t, "Setup database schema", defs[0].Title)
	assert.Equal(t, 1, defs[0].Sequence)
	assert.Empty(t, defs[0].DependsOn)
	assert.Len(t, defs[0].AcceptanceCriteria, 2)

	assert.Equal(t, "Implement user CRUD", defs[1].Title)
	assert.Equal(t, 2, defs[1].Sequence)
	assert.Equal(t, []int{1}, defs[1].DependsOn)
}

func TestParseTaskDefs_Empty(t *testing.T) {
	_, err := parseTaskDefs(nil)
	assert.Error(t, err)

	_, err = parseTaskDefs(json.RawMessage(`{"tasks": []}`))
	assert.Error(t, err)
}

func TestConvertToEntities(t *testing.T) {
	svc := &PlannerService{}
	defs := []TaskDef{
		{
			Title:    "First task",
			Spec:     "Do the first thing",
			Sequence: 1,
			AcceptanceCriteria: []CriteriaDef{
				{ID: "AC1", Description: "It works"},
			},
			Branch: "feature/first",
		},
		{
			Title:     "Second task",
			Spec:      "Do the second thing",
			Sequence:  2,
			DependsOn: []int{1},
			AcceptanceCriteria: []CriteriaDef{
				{ID: "AC1", Description: "It also works"},
			},
		},
	}

	input := PlannerInput{
		PRD:     entities.PRD{ProjectID: "proj-1"},
		Project: entities.Project{DefaultBranch: "main"},
	}
	input.PRD.ID = "prd-1"
	input.Project.ID = "proj-1"

	tasks, err := svc.convertToEntities(defs, input)
	require.NoError(t, err)
	assert.Len(t, tasks, 2)

	// First task: no deps → queued.
	assert.Equal(t, "First task", tasks[0].Title)
	assert.Equal(t, "queued", tasks[0].Status)
	assert.Equal(t, "[]", tasks[0].DependsOn)
	assert.Equal(t, "feature/first", tasks[0].Branch)
	assert.Equal(t, "prd-1", tasks[0].PrdID)
	assert.Equal(t, "proj-1", tasks[0].ProjectID)
	assert.NotEmpty(t, tasks[0].ID)

	// Second task: has deps → blocked.
	assert.Equal(t, "Second task", tasks[1].Title)
	assert.Equal(t, "blocked", tasks[1].Status)
	assert.Equal(t, "main", tasks[1].Branch) // defaulted

	// DependsOn should contain the first task's ID.
	var deps []string
	err = json.Unmarshal([]byte(tasks[1].DependsOn), &deps)
	require.NoError(t, err)
	assert.Equal(t, []string{tasks[0].ID}, deps)
}

func TestConvertToEntities_DependencyMapping(t *testing.T) {
	svc := &PlannerService{}
	defs := []TaskDef{
		{Title: "A", Spec: "a", Sequence: 1},
		{Title: "B", Spec: "b", Sequence: 2, DependsOn: []int{1}},
		{Title: "C", Spec: "c", Sequence: 3, DependsOn: []int{1, 2}},
	}

	input := PlannerInput{
		PRD:     entities.PRD{},
		Project: entities.Project{DefaultBranch: "main"},
	}

	tasks, err := svc.convertToEntities(defs, input)
	require.NoError(t, err)

	// C depends on both A and B.
	var deps []string
	err = json.Unmarshal([]byte(tasks[2].DependsOn), &deps)
	require.NoError(t, err)
	assert.Len(t, deps, 2)
	assert.Contains(t, deps, tasks[0].ID)
	assert.Contains(t, deps, tasks[1].ID)
}

func TestConvertToEntities_UnknownSequence(t *testing.T) {
	svc := &PlannerService{}
	defs := []TaskDef{
		{Title: "A", Spec: "a", Sequence: 1, DependsOn: []int{99}},
	}

	input := PlannerInput{
		PRD:     entities.PRD{},
		Project: entities.Project{DefaultBranch: "main"},
	}

	_, err := svc.convertToEntities(defs, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown sequence 99")
}

func TestConvertToEntities_DefaultBranch(t *testing.T) {
	svc := &PlannerService{}
	defs := []TaskDef{
		{Title: "No branch", Spec: "test", Sequence: 1},
		{Title: "Has branch", Spec: "test", Sequence: 2, Branch: "custom"},
	}

	input := PlannerInput{
		PRD:     entities.PRD{},
		Project: entities.Project{DefaultBranch: "develop"},
	}

	tasks, err := svc.convertToEntities(defs, input)
	require.NoError(t, err)

	assert.Equal(t, "develop", tasks[0].Branch)
	assert.Equal(t, "custom", tasks[1].Branch)
}
