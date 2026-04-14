//go:build integration

package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yolo-hq/yolo/yolotest"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

func TestApproveSuggestion_HappyPath(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	sug := seedSuggestion(t, tx, proj.ID, nil) // status=pending

	result := runAction(t, tx, &SuggestionApproveAction{},
		yolotest.WithEntityName("Suggestion"),
		yolotest.WithEntityID(sug.ID),
		yolotest.WithInput(inputs.ApproveSuggestionInput{}),
	)
	require.True(t, result.Success, "approve should succeed: %s", result.Message)
	assertSuggestionStatus(t, tx, sug.ID, "approved")
}

func TestApproveSuggestion_DenyNotPending(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	sug := seedSuggestion(t, tx, proj.ID, &entities.Suggestion{Status: "approved"})

	result := runAction(t, tx, &SuggestionApproveAction{},
		yolotest.WithEntityName("Suggestion"),
		yolotest.WithEntityID(sug.ID),
		yolotest.WithInput(inputs.ApproveSuggestionInput{}),
	)
	assert.False(t, result.Success, "approve should be denied for non-pending suggestion")
	assert.Equal(t, 403, result.StatusCode)
}

func TestApproveSuggestion_DenyRejected(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	sug := seedSuggestion(t, tx, proj.ID, &entities.Suggestion{Status: "rejected"})

	result := runAction(t, tx, &SuggestionApproveAction{},
		yolotest.WithEntityName("Suggestion"),
		yolotest.WithEntityID(sug.ID),
		yolotest.WithInput(inputs.ApproveSuggestionInput{}),
	)
	assert.False(t, result.Success, "approve should be denied for rejected suggestion")
}