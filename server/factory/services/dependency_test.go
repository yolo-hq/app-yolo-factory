package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

func TestParseDepsJSON(t *testing.T) {
	assert.Nil(t, ParseDeps(""))
	assert.Nil(t, ParseDeps("[]"))
	assert.Nil(t, ParseDeps("invalid"))
	assert.Equal(t, []string{"a", "b"}, ParseDeps(`["a","b"]`))
}

func TestDetectCycle_NoCycle(t *testing.T) {
	tasks := map[string]*entities.Task{
		"A": {DependsOn: "[]"},
		"B": {DependsOn: `["A"]`},
	}
	// C depends on B — no cycle.
	err := detectCycle("C", []string{"B"}, tasks)
	require.NoError(t, err)
}

func TestDetectCycle_SelfDep(t *testing.T) {
	tasks := map[string]*entities.Task{}
	err := detectCycle("A", []string{"A"}, tasks)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cycle detected")
}

func TestDetectCycle_ThreeNodeCycle(t *testing.T) {
	// A→B→C and we try to add C→A.
	tasks := map[string]*entities.Task{
		"A": {DependsOn: `["B"]`},
		"B": {DependsOn: `["C"]`},
	}
	// C depends on A → creates A→B→C→A cycle.
	err := detectCycle("C", []string{"A"}, tasks)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cycle detected")
}

func TestDetectCycle_MissingDep(t *testing.T) {
	// Dep doesn't exist — no cycle (just missing, validated elsewhere).
	tasks := map[string]*entities.Task{}
	err := detectCycle("A", []string{"Z"}, tasks)
	require.NoError(t, err)
}

func TestContainsStr(t *testing.T) {
	assert.True(t, containsStr([]string{"a", "b"}, "a"))
	assert.False(t, containsStr([]string{"a", "b"}, "c"))
	assert.False(t, containsStr(nil, "a"))
}

func TestToJSON_Dep(t *testing.T) {
	assert.Equal(t, `["a","b"]`, ToJSON([]string{"a", "b"}))
	assert.Equal(t, "[]", ToJSON([]string{}))
}
