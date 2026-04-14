//go:build integration

package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yolo-hq/yolo/yolotest"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

func TestApprovePRD_HappyPath(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	prd := seedPRD(t, tx, proj.ID, nil) // status=draft

	result := runAction(t, tx, &PRDApproveAction{},
		yolotest.WithEntityName("PRD"),
		yolotest.WithEntityID(prd.ID),
	)
	require.True(t, result.Success, "approve should succeed: %s", result.Message)
	assertPRDStatus(t, tx, prd.ID, "approved")
}

func TestApprovePRD_DenyNotDraft(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	prd := seedPRD(t, tx, proj.ID, &entities.PRD{Status: "approved"})

	result := runAction(t, tx, &PRDApproveAction{},
		yolotest.WithEntityName("PRD"),
		yolotest.WithEntityID(prd.ID),
	)
	assert.False(t, result.Success, "approve should be denied for non-draft PRD")
	assert.Equal(t, 403, result.StatusCode)
}

func TestApprovePRD_DenyInProgress(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	prd := seedPRD(t, tx, proj.ID, &entities.PRD{Status: "in_progress"})

	result := runAction(t, tx, &PRDApproveAction{},
		yolotest.WithEntityName("PRD"),
		yolotest.WithEntityID(prd.ID),
	)
	assert.False(t, result.Success, "approve should be denied for in_progress PRD")
}