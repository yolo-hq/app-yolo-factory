//go:build integration

package actions_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yolo-hq/yolo/yolotest"

	actionsgen "github.com/yolo-hq/app-yolo-factory/.yolo/gen/adapters/apps/common/factory/actions"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

func TestSubmitPRD_HappyPath(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil) // status=active

	result := runAction(t, tx, &actionsgen.PRDSubmitAction{},
		yolotest.WithEntityName("Project"),
		yolotest.WithEntityID(proj.ID),
		yolotest.WithInput(inputs.SubmitPRDInput{
			ProjectID:          proj.ID,
			Title:              "New Feature",
			Body:               "Implement new feature",
			AcceptanceCriteria: "Feature works as described",
		}),
	)
	require.True(t, result.Success, "submit should succeed for active project: %s", result.Message)
}

func TestSubmitPRD_DenyArchivedProject(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, &entities.Project{Status: "archived"})

	result := runAction(t, tx, &actionsgen.PRDSubmitAction{},
		yolotest.WithEntityName("Project"),
		yolotest.WithEntityID(proj.ID),
		yolotest.WithInput(inputs.SubmitPRDInput{
			ProjectID:          proj.ID,
			Title:              "New Feature",
			Body:               "Implement new feature",
			AcceptanceCriteria: "Feature works",
		}),
	)
	assert.False(t, result.Success, "submit should be denied for archived project")
	assert.Equal(t, 403, result.StatusCode)
}

func TestSubmitPRD_DenyMissingRequiredFields(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)

	result := runAction(t, tx, &actionsgen.PRDSubmitAction{},
		yolotest.WithEntityName("Project"),
		yolotest.WithEntityID(proj.ID),
		yolotest.WithInput(inputs.SubmitPRDInput{
			ProjectID: proj.ID,
			// Missing Title, Body, AcceptanceCriteria
		}),
	)
	assert.False(t, result.Success, "submit should be denied with missing fields")
	assert.Equal(t, 422, result.StatusCode)
}