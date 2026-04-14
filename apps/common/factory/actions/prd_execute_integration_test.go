//go:build integration

package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yolo-hq/yolo/yolotest"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

func TestExecutePRD_HappyPathDraft(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	prd := seedPRD(t, tx, proj.ID, nil) // status=draft

	result := runAction(t, tx, &PRDExecuteAction{},
		yolotest.WithEntityName("PRD"),
		yolotest.WithEntityID(prd.ID),
	)
	require.True(t, result.Success, "execute should succeed for draft: %s", result.Message)
	assertPRDStatus(t, tx, prd.ID, "planning")
}

func TestExecutePRD_HappyPathApproved(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	prd := seedPRD(t, tx, proj.ID, &entities.PRD{Status: "approved"})

	result := runAction(t, tx, &PRDExecuteAction{},
		yolotest.WithEntityName("PRD"),
		yolotest.WithEntityID(prd.ID),
	)
	require.True(t, result.Success, "execute should succeed for approved: %s", result.Message)
	assertPRDStatus(t, tx, prd.ID, "planning")
}

func TestExecutePRD_DenyCompleted(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	prd := seedPRD(t, tx, proj.ID, &entities.PRD{Status: "completed"})

	result := runAction(t, tx, &PRDExecuteAction{},
		yolotest.WithEntityName("PRD"),
		yolotest.WithEntityID(prd.ID),
	)
	assert.False(t, result.Success, "execute should be denied for completed PRD")
	assert.Equal(t, 403, result.StatusCode)
}