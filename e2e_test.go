package main

import (
	"os"
	"testing"

	"github.com/yolo-hq/yolo/yolotest"
)

var srv *yolotest.AppServer

func TestMain(m *testing.M) {
	setup()
	srv = yolotest.Start(nil)
	code := m.Run()
	srv.Close()
	os.Exit(code)
}

func TestE2E_CreateRepo(t *testing.T) {
	resp := srv.POST("/api/v1/factory/repos", map[string]any{
		"name": "test-repo",
		"url":  "https://github.com/test/repo",
	})
	resp.AssertStatus(t, 201)
	resp.AssertSuccess(t)
}

func TestE2E_CreateRepo_Duplicate(t *testing.T) {
	srv.POST("/api/v1/factory/repos", map[string]any{
		"name": "dup-repo",
		"url":  "https://github.com/test/dup",
	})
	resp := srv.POST("/api/v1/factory/repos", map[string]any{
		"name": "dup-repo",
		"url":  "https://github.com/test/dup",
	})
	resp.AssertError(t)
}

func TestE2E_ListRepos(t *testing.T) {
	resp := srv.GET("/api/v1/factory/repos?fields.repo=id,name")
	resp.AssertStatus(t, 200)
	resp.AssertSuccess(t)

	repos := resp.Field("data.repos")
	if !repos.Exists() {
		t.Fatal("expected repos in response")
	}
}

func TestE2E_ListRepos_IDOnly(t *testing.T) {
	resp := srv.GET("/api/v1/factory/repos")
	resp.AssertSuccess(t)
}

func TestE2E_CreateTask(t *testing.T) {
	resp := srv.POST("/api/v1/factory/tasks", map[string]any{
		"repoId":   "test-repo-id",
		"title":    "E2E Test Task",
		"priority": 5,
	})
	resp.AssertStatus(t, 201)
	resp.AssertSuccess(t)
}

func TestE2E_ListTasks(t *testing.T) {
	resp := srv.GET("/api/v1/factory/tasks?fields.task=id,title,status,priority")
	resp.AssertStatus(t, 200)
	resp.AssertSuccess(t)

	tasks := resp.Field("data.tasks")
	if !tasks.Exists() {
		t.Fatal("expected tasks in response")
	}
}
