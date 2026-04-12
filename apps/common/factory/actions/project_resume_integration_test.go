package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yolo-hq/yolo/yolotest"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

func TestResumeProject_HappyPath(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, &entities.Project{Status: "paused"})

	result := runAction(t, tx, &ResumeProjectAction{},
		yolotest.WithEntityName("Project"),
		yolotest.WithEntityID(proj.ID),
	)
	require.True(t, result.Success, "resume should succeed: %s", result.Message)
	assertProjectStatus(t, tx, proj.ID, "active")
}

func TestResumeProject_DenyNotPaused(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil) // status=active

	result := runAction(t, tx, &ResumeProjectAction{},
		yolotest.WithEntityName("Project"),
		yolotest.WithEntityID(proj.ID),
	)
	assert.False(t, result.Success, "resume should be denied for active project")
	assert.Equal(t, 403, result.StatusCode)
}

func TestResumeProject_DenyArchived(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, &entities.Project{Status: "archived"})

	result := runAction(t, tx, &ResumeProjectAction{},
		yolotest.WithEntityName("Project"),
		yolotest.WithEntityID(proj.ID),
	)
	assert.False(t, result.Success, "resume should be denied for archived project")
}
