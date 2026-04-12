package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yolo-hq/yolo/yolotest"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
)

func TestAcknowledgeInsight_HappyPath(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	ins := seedInsight(t, tx, proj.ID, nil) // status=pending

	result := runAction(t, tx, &AcknowledgeInsightAction{},
		yolotest.WithEntityName("Insight"),
		yolotest.WithEntityID(ins.ID),
	)
	require.True(t, result.Success, "acknowledge should succeed: %s", result.Message)
	assertInsightStatus(t, tx, ins.ID, "acknowledged")
}

func TestAcknowledgeInsight_DenyNotPending(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	ins := seedInsight(t, tx, proj.ID, &entities.Insight{Status: "acknowledged"})

	result := runAction(t, tx, &AcknowledgeInsightAction{},
		yolotest.WithEntityName("Insight"),
		yolotest.WithEntityID(ins.ID),
	)
	assert.False(t, result.Success, "acknowledge should be denied for non-pending insight")
	assert.Equal(t, 403, result.StatusCode)
}

func TestAcknowledgeInsight_DenyDismissed(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	ins := seedInsight(t, tx, proj.ID, &entities.Insight{Status: "dismissed"})

	result := runAction(t, tx, &AcknowledgeInsightAction{},
		yolotest.WithEntityName("Insight"),
		yolotest.WithEntityID(ins.ID),
	)
	assert.False(t, result.Success, "acknowledge should be denied for dismissed insight")
}
