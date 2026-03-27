package entities_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

func TestRepoImplementsInterfaces(t *testing.T) {
	var r entities.Repo
	var _ entity.Entity = r
	var _ entity.Named = r
	var _ entity.Timestamped = r
	var _ entity.SoftDeletable = r

	assert.Equal(t, "repos", r.TableName())
	assert.Equal(t, "Repo", r.EntityName())
}

func TestRepoBaseEntity(t *testing.T) {
	r := entities.Repo{}
	r.ID = "test-123"
	assert.Equal(t, "test-123", r.GetID())
	assert.False(t, r.IsDeleted())
}

func TestTaskImplementsInterfaces(t *testing.T) {
	var task entities.Task
	var _ entity.Entity = task
	var _ entity.Named = task
	var _ entity.Timestamped = task
	var _ entity.SoftDeletable = task

	assert.Equal(t, "tasks", task.TableName())
	assert.Equal(t, "Task", task.EntityName())
}

func TestTaskBaseEntity(t *testing.T) {
	task := entities.Task{}
	task.ID = "task-123"
	assert.Equal(t, "task-123", task.GetID())
	assert.False(t, task.IsDeleted())
}
