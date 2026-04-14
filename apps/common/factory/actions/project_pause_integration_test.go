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

func TestPauseProject_HappyPath(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil) // status=active

	result := runAction(t, tx, &actionsgen.ProjectPauseAction{},
		yolotest.WithEntityName("Project"),
		yolotest.WithEntityID(proj.ID),
	)
	require.True(t, result.Success, "pause should succeed: %s", result.Message)
	assertProjectStatus(t, tx, proj.ID, "paused")
}

func TestPauseProject_DenyNotActive(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, &entities.Project{Status: "paused"})

	result := runAction(t, tx, &actionsgen.ProjectPauseAction{},
		yolotest.WithEntityName("Project"),
		yolotest.WithEntityID(proj.ID),
	)
	assert.False(t, result.Success, "pause should be denied for already paused project")
	assert.Equal(t, 403, result.StatusCode)
}

func TestPauseProject_DenyArchived(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, &entities.Project{Status: "archived"})

	result := runAction(t, tx, &actionsgen.ProjectPauseAction{},
		yolotest.WithEntityName("Project"),
		yolotest.WithEntityID(proj.ID),
	)
	assert.False(t, result.Success, "pause should be denied for archived project")
}