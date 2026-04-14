//go:build integration

package actions_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yolo-hq/yolo/yolotest"

	actionsgen "github.com/yolo-hq/app-yolo-factory/.yolo/gen/adapters/apps/common/factory/actions"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

func TestUpdateProject_HappyPath(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	newName := "Updated Name " + newID()

	result := runAction(t, tx, &actionsgen.ProjectUpdateAction{},
		yolotest.WithEntityName("Project"),
		yolotest.WithEntityID(proj.ID),
		yolotest.WithInput(inputs.UpdateProjectInput{
			Name: &newName,
		}),
	)
	require.True(t, result.Success, "update should succeed: %s", result.Message)

	var got struct{ Name string `bun:"name"` }
	err := tx.NewSelect().TableExpr("factory_projects").
		ColumnExpr("name").Where("id = ?", proj.ID).
		Scan(context.Background(), &got)
	require.NoError(t, err)
	assert.Equal(t, newName, got.Name)
}

func TestUpdateProject_UpdateModel(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	model := "opus"

	result := runAction(t, tx, &actionsgen.ProjectUpdateAction{},
		yolotest.WithEntityName("Project"),
		yolotest.WithEntityID(proj.ID),
		yolotest.WithInput(inputs.UpdateProjectInput{
			DefaultModel: &model,
		}),
	)
	require.True(t, result.Success, "update model should succeed: %s", result.Message)

	var got struct{ Model string `bun:"default_model"` }
	err := tx.NewSelect().TableExpr("factory_projects").
		ColumnExpr("default_model").Where("id = ?", proj.ID).
		Scan(context.Background(), &got)
	require.NoError(t, err)
	assert.Equal(t, "opus", got.Model)
}