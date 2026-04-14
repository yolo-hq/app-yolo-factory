//go:build integration

package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yolo-hq/yolo/yolotest"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

func TestRetryTask_HappyPath(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	prd := seedPRD(t, tx, proj.ID, nil)
	task := seedTask(t, tx, proj.ID, prd.ID, &entities.Task{Status: "failed"})

	result := runAction(t, tx, &RetryTaskAction{},
		yolotest.WithEntityName("Task"),
		yolotest.WithEntityID(task.ID),
		yolotest.WithInput(inputs.RetryTaskInput{}),
	)
	require.True(t, result.Success, "retry should succeed: %s", result.Message)
	assertTaskStatus(t, tx, task.ID, "queued")
}

func TestRetryTask_HappyPathWithModelOverride(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	prd := seedPRD(t, tx, proj.ID, nil)
	task := seedTask(t, tx, proj.ID, prd.ID, &entities.Task{Status: "failed"})

	result := runAction(t, tx, &RetryTaskAction{},
		yolotest.WithEntityName("Task"),
		yolotest.WithEntityID(task.ID),
		yolotest.WithInput(inputs.RetryTaskInput{Model: "opus"}),
	)
	require.True(t, result.Success, "retry with model override should succeed: %s", result.Message)
	assertTaskStatus(t, tx, task.ID, "queued")
}

func TestRetryTask_DenyNotFailed(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	prd := seedPRD(t, tx, proj.ID, nil)
	task := seedTask(t, tx, proj.ID, prd.ID, nil) // status=queued

	result := runAction(t, tx, &RetryTaskAction{},
		yolotest.WithEntityName("Task"),
		yolotest.WithEntityID(task.ID),
		yolotest.WithInput(inputs.RetryTaskInput{}),
	)
	assert.False(t, result.Success, "retry should be denied for non-failed task")
	assert.Equal(t, 403, result.StatusCode)
}