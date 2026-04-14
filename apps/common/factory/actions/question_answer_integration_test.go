//go:build integration

package actions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yolo-hq/yolo/yolotest"

	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/entities"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

func TestAnswerQuestion_HappyPath(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	prd := seedPRD(t, tx, proj.ID, nil)
	task := seedTask(t, tx, proj.ID, prd.ID, nil)
	run := seedRun(t, tx, task.ID, nil)
	q := seedQuestion(t, tx, task.ID, run.ID, nil) // status=open

	result := runAction(t, tx, &AnswerQuestionAction{},
		yolotest.WithEntityName("Question"),
		yolotest.WithEntityID(q.ID),
		yolotest.WithInput(inputs.AnswerQuestionInput{Answer: "Use the existing pattern"}),
	)
	require.True(t, result.Success, "answer should succeed: %s", result.Message)
	assertQuestionStatus(t, tx, q.ID, "answered")

	var got struct {
		Answer string `bun:"answer"`
	}
	err := tx.NewSelect().TableExpr("factory_questions").
		ColumnExpr("answer").Where("id = ?", q.ID).
		Scan(context.Background(), &got)
	require.NoError(t, err)
	assert.Equal(t, "Use the existing pattern", got.Answer)
}

func TestAnswerQuestion_DenyNotOpen(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	prd := seedPRD(t, tx, proj.ID, nil)
	task := seedTask(t, tx, proj.ID, prd.ID, nil)
	run := seedRun(t, tx, task.ID, nil)
	q := seedQuestion(t, tx, task.ID, run.ID, &entities.Question{Status: "answered"})

	result := runAction(t, tx, &AnswerQuestionAction{},
		yolotest.WithEntityName("Question"),
		yolotest.WithEntityID(q.ID),
		yolotest.WithInput(inputs.AnswerQuestionInput{Answer: "Answer attempt"}),
	)
	assert.False(t, result.Success, "answer should be denied for non-open question")
	assert.Equal(t, 403, result.StatusCode)
}

func TestAnswerQuestion_DenyMissingAnswer(t *testing.T) {
	tx := dbTx(t)
	proj := seedProject(t, tx, nil)
	prd := seedPRD(t, tx, proj.ID, nil)
	task := seedTask(t, tx, proj.ID, prd.ID, nil)
	run := seedRun(t, tx, task.ID, nil)
	q := seedQuestion(t, tx, task.ID, run.ID, nil)

	result := runAction(t, tx, &AnswerQuestionAction{},
		yolotest.WithEntityName("Question"),
		yolotest.WithEntityID(q.ID),
		yolotest.WithInput(inputs.AnswerQuestionInput{
			// Missing Answer (required)
		}),
	)
	assert.False(t, result.Success, "answer should fail without answer text")
	assert.Equal(t, 422, result.StatusCode)
}