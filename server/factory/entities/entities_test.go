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
