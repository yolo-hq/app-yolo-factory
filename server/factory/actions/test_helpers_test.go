package actions

import (
	"context"

	"github.com/yolo-hq/yolo/core/entity"

	"github.com/yolo-hq/app-yolo-factory/server/factory/entities"
)

// mockTaskReader is a test mock for entity.ReadRepository[entities.Task].
type mockTaskReader struct {
	tasks map[string]*entities.Task
}

func (m *mockTaskReader) FindOne(_ context.Context, opts entity.FindOneOptions) (*entities.Task, error) {
	if t, ok := m.tasks[opts.ID]; ok {
		return t, nil
	}
	return nil, nil
}

func (m *mockTaskReader) FindMany(_ context.Context, opts entity.FindOptions) (entity.FindResult[entities.Task], error) {
	var result []entities.Task
	for _, t := range m.tasks {
		// Apply status filter if present
		for _, f := range opts.Filters {
			if f.Field == "status" && f.Value != t.Status {
				goto skip
			}
		}
		result = append(result, *t)
	skip:
	}
	return entity.FindResult[entities.Task]{Data: result}, nil
}

func (m *mockTaskReader) Count(_ context.Context, _ entity.CountOptions) (int, error) {
	return len(m.tasks), nil
}

func newTask(id, status, dependsOn string) *entities.Task {
	t := &entities.Task{Status: status, DependsOn: dependsOn}
	t.ID = id
	return t
}
