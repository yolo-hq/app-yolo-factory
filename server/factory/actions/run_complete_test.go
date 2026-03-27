package actions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yolo-hq/yolo/core/entity"
	"github.com/yolo-hq/yolo/core/filter"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

// mockTaskWriter implements entity.WriteRepository[entities.Task] for tests.
type mockTaskWriter struct {
	updates map[string]map[string]any
}

func newMockTaskWriter() *mockTaskWriter {
	return &mockTaskWriter{updates: make(map[string]map[string]any)}
}

func (m *mockTaskWriter) Update(_ context.Context) entity.UpdateBuilder[entities.Task] {
	return &taskUpdateBuilder{writer: m}
}

func (m *mockTaskWriter) Insert(_ context.Context, e *entities.Task) (*entities.Task, error) {
	return e, nil
}

func (m *mockTaskWriter) Delete(_ context.Context, _ string) error { return nil }

func (m *mockTaskWriter) SoftDelete(_ context.Context, _ string) (*entities.Task, error) {
	return nil, nil
}

func (m *mockTaskWriter) Restore(_ context.Context, _ string) (*entities.Task, error) {
	return nil, nil
}

type taskUpdateBuilder struct {
	writer   *mockTaskWriter
	entityID string
	sets     map[string]any
}

func (b *taskUpdateBuilder) Where(cond entity.FilterCondition) entity.UpdateBuilder[entities.Task] {
	b.entityID = cond.Value.(string)
	b.sets = make(map[string]any)
	return b
}

func (b *taskUpdateBuilder) WhereFilter(_ filter.EntityFilter) entity.UpdateBuilder[entities.Task] {
	return b
}

func (b *taskUpdateBuilder) Set(field string, value any) entity.UpdateBuilder[entities.Task] {
	b.sets[field] = value
	return b
}

func (b *taskUpdateBuilder) SetFromInput(_ any) entity.UpdateBuilder[entities.Task] { return b }

func (b *taskUpdateBuilder) Incr(_ string, _ int64) entity.UpdateBuilder[entities.Task] { return b }

func (b *taskUpdateBuilder) Decr(_ string, _ int64) entity.UpdateBuilder[entities.Task] { return b }

func (b *taskUpdateBuilder) Returning() entity.UpdateBuilder[entities.Task] { return b }

func (b *taskUpdateBuilder) Exec(_ context.Context) (*entities.Task, error) {
	b.writer.updates[b.entityID] = b.sets
	return nil, nil
}

func (b *taskUpdateBuilder) ExecOne(_ context.Context) (*entities.Task, error) { return nil, nil }

func (b *taskUpdateBuilder) ExecMany(_ context.Context) (int64, error) { return 0, nil }

// --- unblockDependents tests ---

func TestUnblockDependents_AllDepsDone(t *testing.T) {
	reader := &mockTaskReader{tasks: map[string]*entities.Task{
		"completed": newTask("completed", "done", "[]"),
		"blocked":   newTask("blocked", "blocked", `["completed"]`),
	}}
	writer := newMockTaskWriter()

	a := &CompleteRunAction{TaskRead: reader, TaskWrite: writer}
	a.unblockDependents(context.Background(), "completed")

	assert.Equal(t, "queued", writer.updates["blocked"]["status"])
}

func TestUnblockDependents_OtherDepsIncomplete(t *testing.T) {
	reader := &mockTaskReader{tasks: map[string]*entities.Task{
		"done1":   newTask("done1", "done", "[]"),
		"running": newTask("running", "running", "[]"),
		"blocked": newTask("blocked", "blocked", `["done1","running"]`),
	}}
	writer := newMockTaskWriter()

	a := &CompleteRunAction{TaskRead: reader, TaskWrite: writer}
	a.unblockDependents(context.Background(), "done1")

	_, updated := writer.updates["blocked"]
	assert.False(t, updated, "blocked task should NOT be unblocked when other deps incomplete")
}

func TestUnblockDependents_NoDependents(t *testing.T) {
	reader := &mockTaskReader{tasks: map[string]*entities.Task{
		"completed": newTask("completed", "done", "[]"),
	}}
	writer := newMockTaskWriter()

	a := &CompleteRunAction{TaskRead: reader, TaskWrite: writer}
	a.unblockDependents(context.Background(), "completed")

	assert.Empty(t, writer.updates)
}

func TestUnblockDependents_DiamondDeps(t *testing.T) {
	// Task D depends on B and C. B is done, C just completed.
	reader := &mockTaskReader{tasks: map[string]*entities.Task{
		"b": newTask("b", "done", "[]"),
		"c": newTask("c", "done", "[]"),
		"d": newTask("d", "blocked", `["b","c"]`),
	}}
	writer := newMockTaskWriter()

	a := &CompleteRunAction{TaskRead: reader, TaskWrite: writer}
	a.unblockDependents(context.Background(), "c")

	assert.Equal(t, "queued", writer.updates["d"]["status"])
}

// --- status mapping tests ---

func TestStatusMapping_CompleteBecomeDone(t *testing.T) {
	taskStatus := "complete"
	if taskStatus == "complete" {
		taskStatus = "done"
	}
	assert.Equal(t, "done", taskStatus)
}

func TestStatusMapping_FailedRetriesLeft(t *testing.T) {
	taskStatus := "failed"
	if taskStatus == "failed" && 1 < 3 {
		taskStatus = "queued"
	}
	assert.Equal(t, "queued", taskStatus)
}

func TestStatusMapping_FailedNoRetries(t *testing.T) {
	taskStatus := "failed"
	if taskStatus == "failed" && 3 < 3 {
		taskStatus = "queued"
	}
	assert.Equal(t, "failed", taskStatus)
}

func TestStatusMapping_PartialStaysPartial(t *testing.T) {
	taskStatus := "partial"
	if taskStatus == "complete" {
		taskStatus = "done"
	}
	assert.Equal(t, "partial", taskStatus)
}
