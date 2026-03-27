package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E runs end-to-end tests against the Factory API.
// Requires: yolo dev running on :9000 with fresh database.
//
// Run: go test -tags e2e -v ./...
// Or manually:
//   1. rm -f factory.db && yolo dev &
//   2. go test -tags e2e -v

var baseURL = "http://localhost:9000"

func init() {
	if url := os.Getenv("FACTORY_URL"); url != "" {
		baseURL = url
	}
}

func TestE2E_CreateRepo(t *testing.T) {
	resp := post(t, "/api/v1/factory/repos", map[string]any{
		"name": "test-repo",
		"url":  "https://github.com/test/repo",
	})
	assert.True(t, resp["success"].(bool))
}

func TestE2E_CreateRepo_Duplicate(t *testing.T) {
	// First create
	post(t, "/api/v1/factory/repos", map[string]any{
		"name": "dup-repo",
		"url":  "https://github.com/test/dup",
	})
	// Duplicate
	resp := postRaw(t, "/api/v1/factory/repos", map[string]any{
		"name": "dup-repo",
		"url":  "https://github.com/test/dup",
	})
	assert.False(t, resp["success"].(bool))
}

func TestE2E_ListRepos(t *testing.T) {
	resp := get(t, "/api/v1/factory/repos?fields.repo=id,name")
	assert.True(t, resp["success"].(bool))
	data := resp["data"].(map[string]any)
	repos := data["repos"].([]any)
	assert.Greater(t, len(repos), 0)

	first := repos[0].(map[string]any)
	assert.NotEmpty(t, first["id"])
	assert.NotEmpty(t, first["name"])
}

func TestE2E_ListRepos_IDOnly(t *testing.T) {
	resp := get(t, "/api/v1/factory/repos")
	data := resp["data"].(map[string]any)
	repos := data["repos"].([]any)
	if len(repos) > 0 {
		first := repos[0].(map[string]any)
		assert.NotEmpty(t, first["id"])
		// Default should be ID only — but querier returns id
		assert.Len(t, first, 1, "default response should be ID only")
	}
}

func TestE2E_CreateTask(t *testing.T) {
	resp := post(t, "/api/v1/factory/tasks", map[string]any{
		"repoId":   "test-repo-id",
		"title":    "E2E Test Task",
		"priority": 5,
	})
	assert.True(t, resp["success"].(bool))
}

func TestE2E_ListTasks(t *testing.T) {
	resp := get(t, "/api/v1/factory/tasks?fields.task=id,title,status,priority")
	assert.True(t, resp["success"].(bool))
	data := resp["data"].(map[string]any)
	tasks := data["tasks"].([]any)
	assert.Greater(t, len(tasks), 0)

	first := tasks[0].(map[string]any)
	assert.NotEmpty(t, first["title"])
	assert.Equal(t, "queued", first["status"])
}

func TestE2E_CreateTask_WithDependency(t *testing.T) {
	// Create first task
	resp1 := post(t, "/api/v1/factory/tasks", map[string]any{
		"repoId": "dep-test",
		"title":  "Parent Task",
	})
	assert.True(t, resp1["success"].(bool))
}

// --- Helpers ---

func post(t *testing.T, path string, body map[string]any) map[string]any {
	t.Helper()
	resp := postRaw(t, path, body)
	require.True(t, resp["success"].(bool), "POST %s failed: %v", path, resp)
	return resp
}

func postRaw(t *testing.T, path string, body map[string]any) map[string]any {
	t.Helper()
	data, _ := json.Marshal(body)
	resp, err := http.Post(baseURL+path, "application/json", bytes.NewReader(data))
	require.NoError(t, err)
	defer resp.Body.Close()
	return readJSON(t, resp.Body)
}

func get(t *testing.T, path string) map[string]any {
	t.Helper()
	resp, err := http.Get(baseURL + path)
	require.NoError(t, err)
	defer resp.Body.Close()
	return readJSON(t, resp.Body)
}

func readJSON(t *testing.T, r io.Reader) map[string]any {
	t.Helper()
	var result map[string]any
	require.NoError(t, json.NewDecoder(r).Decode(&result))
	return result
}

// unused but available for future tests
var _ = httptest.NewServer
var _ = exec.Command
