package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yolo-hq/yolo/yolotest"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

func TestApplyInsight_HappyPath(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	ins := seedInsight(t, tx, proj.ID, &entities.Insight{Status: "acknowledged"})

	result := runAction(t, tx, &ApplyInsightAction{},
		yolotest.WithEntityName("Insight"),
		yolotest.WithEntityID(ins.ID),
	)
	require.True(t, result.Success, "apply should succeed: %s", result.Message)
	assertInsightStatus(t, tx, ins.ID, "applied")
}

func TestApplyInsight_DenyPending(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	ins := seedInsight(t, tx, proj.ID, nil) // status=pending

	result := runAction(t, tx, &ApplyInsightAction{},
		yolotest.WithEntityName("Insight"),
		yolotest.WithEntityID(ins.ID),
	)
	assert.False(t, result.Success, "apply should be denied for pending insight")
	assert.Equal(t, 403, result.StatusCode)
}

func TestApplyInsight_DenyAlreadyApplied(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	ins := seedInsight(t, tx, proj.ID, &entities.Insight{Status: "applied"})

	result := runAction(t, tx, &ApplyInsightAction{},
		yolotest.WithEntityName("Insight"),
		yolotest.WithEntityID(ins.ID),
	)
	assert.False(t, result.Success, "apply should be denied for already applied insight")
}
