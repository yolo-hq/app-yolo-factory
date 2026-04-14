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

func TestRejectSuggestion_HappyPath(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	sug := seedSuggestion(t, tx, proj.ID, nil) // status=pending

	result := runAction(t, tx, &actionsgen.SuggestionRejectAction{},
		yolotest.WithEntityName("Suggestion"),
		yolotest.WithEntityID(sug.ID),
		yolotest.WithInput(inputs.RejectSuggestionInput{Reason: "Not aligned with roadmap"}),
	)
	require.True(t, result.Success, "reject should succeed: %s", result.Message)
	assertSuggestionStatus(t, tx, sug.ID, "rejected")
}

func TestRejectSuggestion_DenyNotPending(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	sug := seedSuggestion(t, tx, proj.ID, &entities.Suggestion{Status: "rejected"})

	result := runAction(t, tx, &actionsgen.SuggestionRejectAction{},
		yolotest.WithEntityName("Suggestion"),
		yolotest.WithEntityID(sug.ID),
		yolotest.WithInput(inputs.RejectSuggestionInput{Reason: "Already rejected"}),
	)
	assert.False(t, result.Success, "reject should be denied for non-pending suggestion")
	assert.Equal(t, 403, result.StatusCode)
}

func TestRejectSuggestion_DenyMissingReason(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	sug := seedSuggestion(t, tx, proj.ID, nil)

	result := runAction(t, tx, &actionsgen.SuggestionRejectAction{},
		yolotest.WithEntityName("Suggestion"),
		yolotest.WithEntityID(sug.ID),
		yolotest.WithInput(inputs.RejectSuggestionInput{
			// Missing Reason (required)
		}),
	)
	assert.False(t, result.Success, "reject should fail without reason")
	assert.Equal(t, 422, result.StatusCode)
}