//go:build integration

package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yolo-hq/yolo/yolotest"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

func TestDismissInsight_HappyPathPending(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	ins := seedInsight(t, tx, proj.ID, nil) // status=pending

	result := runAction(t, tx, &DismissInsightAction{},
		yolotest.WithEntityName("Insight"),
		yolotest.WithEntityID(ins.ID),
	)
	require.True(t, result.Success, "dismiss pending should succeed: %s", result.Message)
	assertInsightStatus(t, tx, ins.ID, "dismissed")
}

func TestDismissInsight_HappyPathAcknowledged(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	ins := seedInsight(t, tx, proj.ID, &entities.Insight{Status: "acknowledged"})

	result := runAction(t, tx, &DismissInsightAction{},
		yolotest.WithEntityName("Insight"),
		yolotest.WithEntityID(ins.ID),
	)
	require.True(t, result.Success, "dismiss acknowledged should succeed: %s", result.Message)
	assertInsightStatus(t, tx, ins.ID, "dismissed")
}

func TestDismissInsight_DenyAlreadyApplied(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	ins := seedInsight(t, tx, proj.ID, &entities.Insight{Status: "applied"})

	result := runAction(t, tx, &DismissInsightAction{},
		yolotest.WithEntityName("Insight"),
		yolotest.WithEntityID(ins.ID),
	)
	assert.False(t, result.Success, "dismiss should be denied for applied insight")
	assert.Equal(t, 403, result.StatusCode)
}