package economyhooks

import (
	"context"
	"errors"
)

type TaskFilter struct {
	Lang     string
	Category string
	Keys     []string
}

type (
	TaskLister  func(ctx context.Context, userID string, f TaskFilter) ([]map[string]any, error)
	TaskClaimer func(ctx context.Context, userID, idOrKey, lang string) (map[string]any, error)
)

var (
	ErrNoTasks = errors.New("economy tasks are not available: the economy plugin is not loaded")

	taskLister  TaskLister
	taskClaimer TaskClaimer
)

func SetTasks(list TaskLister, claim TaskClaimer) {
	mu.Lock()
	defer mu.Unlock()
	taskLister, taskClaimer = list, claim
}

func TasksAvailable() bool {
	mu.RLock()
	defer mu.RUnlock()
	return taskLister != nil && taskClaimer != nil
}

func ListTasks(ctx context.Context, userID string, f TaskFilter) ([]map[string]any, error) {
	mu.RLock()
	fn := taskLister
	mu.RUnlock()
	if fn == nil {
		return nil, ErrNoTasks
	}
	return fn(ctx, userID, f)
}

func ClaimTask(ctx context.Context, userID, idOrKey, lang string) (map[string]any, error) {
	mu.RLock()
	fn := taskClaimer
	mu.RUnlock()
	if fn == nil {
		return nil, ErrNoTasks
	}
	return fn(ctx, userID, idOrKey, lang)
}
