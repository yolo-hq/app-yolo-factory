//go:build integration

package actions_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yolo-hq/yolo/yolotest"

	actionsgen "github.com/yolo-hq/app-yolo-factory/.yolo/gen/adapters/apps/common/factory/actions"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

func TestArchiveProject_HappyPath(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil) // status=active

	result := runAction(t, tx, &actionsgen.ProjectArchiveAction{},
		yolotest.WithEntityName("Project"),
		yolotest.WithEntityID(proj.ID),
	)
	require.True(t, result.Success, "archive should succeed: %s", result.Message)
	assertProjectStatus(t, tx, proj.ID, "archived")
}

func TestArchiveProject_DenyAlreadyArchived(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, &entities.Project{Status: "archived"})

	result := runAction(t, tx, &actionsgen.ProjectArchiveAction{},
		yolotest.WithEntityName("Project"),
		yolotest.WithEntityID(proj.ID),
	)
	assert.False(t, result.Success, "archive should be denied for already archived project")
	assert.Equal(t, 403, result.StatusCode)
}

func TestArchiveProject_FromPaused(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, &entities.Project{Status: "paused"})

	result := runAction(t, tx, &actionsgen.ProjectArchiveAction{},
		yolotest.WithEntityName("Project"),
		yolotest.WithEntityID(proj.ID),
	)
	require.True(t, result.Success, "archive paused project should succeed: %s", result.Message)
	assertProjectStatus(t, tx, proj.ID, "archived")
}