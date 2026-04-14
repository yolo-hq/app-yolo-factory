//go:build integration

package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yolo-hq/yolo/yolotest"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

func TestCancelTask_HappyPathQueued(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	prd := seedPRD(t, tx, proj.ID, nil)
	task := seedTask(t, tx, proj.ID, prd.ID, nil) // status=queued

	result := runAction(t, tx, &CancelTaskAction{},
		yolotest.WithEntityName("Task"),
		yolotest.WithEntityID(task.ID),
	)
	require.True(t, result.Success, "cancel queued task should succeed: %s", result.Message)
	assertTaskStatus(t, tx, task.ID, "cancelled")
}

func TestCancelTask_HappyPathRunning(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	prd := seedPRD(t, tx, proj.ID, nil)
	task := seedTask(t, tx, proj.ID, prd.ID, &entities.Task{Status: "running"})

	result := runAction(t, tx, &CancelTaskAction{},
		yolotest.WithEntityName("Task"),
		yolotest.WithEntityID(task.ID),
	)
	require.True(t, result.Success, "cancel running task should succeed: %s", result.Message)
	assertTaskStatus(t, tx, task.ID, "cancelled")
}

func TestCancelTask_DenyDone(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	prd := seedPRD(t, tx, proj.ID, nil)
	task := seedTask(t, tx, proj.ID, prd.ID, &entities.Task{Status: "done"})

	result := runAction(t, tx, &CancelTaskAction{},
		yolotest.WithEntityName("Task"),
		yolotest.WithEntityID(task.ID),
	)
	assert.False(t, result.Success, "cancel should be denied for done task")
	assert.Equal(t, 403, result.StatusCode)
}

func TestCancelTask_DenyAlreadyCancelled(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	prd := seedPRD(t, tx, proj.ID, nil)
	task := seedTask(t, tx, proj.ID, prd.ID, &entities.Task{Status: "cancelled"})

	result := runAction(t, tx, &CancelTaskAction{},
		yolotest.WithEntityName("Task"),
		yolotest.WithEntityID(task.ID),
	)
	assert.False(t, result.Success, "cancel should be denied for already cancelled task")
}