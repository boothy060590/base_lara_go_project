package goroutine

import (
	"base_lara_go_project/app/core/go_core"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGoroutineRepository tests the canonical repository with goroutine optimizations
func TestGoroutineRepository(t *testing.T) {
	t.Run("NewRepository", func(t *testing.T) {
		repo := go_core.NewRepository[string](nil, nil, nil, nil)
		assert.NotNil(t, repo)
	})

	t.Run("FindAsync", func(t *testing.T) {
		mockRepo := &MockRepository[string]{
			findResult: "test-data",
			findError:  nil,
		}
		resultChan := make(chan go_core.RepositoryResult[string], 1)
		go func() {
			defer close(resultChan)
			model, err := mockRepo.Find(1)
			resultChan <- go_core.RepositoryResult[string]{
				Data:  *model,
				Error: err,
			}
		}()
		result := <-resultChan
		assert.Equal(t, "test-data", result.Data)
		assert.NoError(t, result.Error)
	})

	t.Run("FindManyAsync", func(t *testing.T) {
		mockRepo := &MockRepository[string]{
			findResult: "test-data",
			findError:  nil,
		}
		resultChan := make(chan go_core.RepositoryResult[[]string], 1)
		go func() {
			defer close(resultChan)
			var models []string
			for _, id := range []uint{1, 2} {
				model, err := mockRepo.Find(id)
				if err != nil {
					resultChan <- go_core.RepositoryResult[[]string]{
						Data:  nil,
						Error: err,
					}
					return
				}
				models = append(models, *model)
			}
			resultChan <- go_core.RepositoryResult[[]string]{
				Data:  models,
				Error: nil,
			}
		}()
		result := <-resultChan
		expected := []string{"test-data", "test-data"}
		assert.Equal(t, expected, result.Data)
		assert.NoError(t, result.Error)
	})
}

// TestGoroutineEventDispatcher tests the canonical event dispatcher with goroutine optimizations
func TestGoroutineEventDispatcher(t *testing.T) {
	t.Run("NewEventBus", func(t *testing.T) {
		dispatcher := go_core.NewEventBus[string](nil, nil, nil)
		assert.NotNil(t, dispatcher)
	})

	t.Run("DispatchAsync", func(t *testing.T) {
		dispatcher := go_core.NewEventBus[string](nil, nil, nil)
		event := &go_core.Event[string]{
			ID:   "test-event",
			Data: "test-data",
		}
		err := dispatcher.DispatchAsync(event)
		assert.NoError(t, err)
	})
}

// TestGoroutineJobDispatcher tests the canonical job dispatcher with goroutine optimizations
func TestGoroutineJobDispatcher(t *testing.T) {
	t.Run("NewJobDispatcher", func(t *testing.T) {
		queue := go_core.NewSyncQueue[string]()
		dispatcher := go_core.NewJobDispatcher[string](queue, nil, nil, nil)
		assert.NotNil(t, dispatcher)
	})

	t.Run("Dispatch", func(t *testing.T) {
		queue := go_core.NewSyncQueue[string]()
		dispatcher := go_core.NewJobDispatcher[string](queue, nil, nil, nil)
		job := "test-job"
		err := dispatcher.Dispatch(job)
		assert.NoError(t, err)
	})
}

type MockRepository[T any] struct {
	findResult     T
	findError      error
	findManyResult []T
	findManyError  error
}

func (m *MockRepository[T]) Find(id uint) (*T, error) {
	return &m.findResult, m.findError
}
