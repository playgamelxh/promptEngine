package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusStopped   TaskStatus = "stopped"
)

type Task struct {
	ID        string      `json:"id"`
	Status    TaskStatus  `json:"status"`
	Progress  int         `json:"progress"`
	Total     int         `json:"total"`
	Message   string      `json:"message"`
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	
	mu     sync.RWMutex
	cancel context.CancelFunc
}

type TaskManager struct {
	tasks sync.Map
}

var GlobalTaskManager = &TaskManager{}

func (tm *TaskManager) StartTask(total int, runFunc func(ctx context.Context, updateProgress func(current int, msg string) error) error) string {
	id := uuid.New().String()
	ctx, cancel := context.WithCancel(context.Background())

	task := &Task{
		ID:        id,
		Status:    TaskStatusPending,
		Total:     total,
		CreatedAt: time.Now(),
		cancel:    cancel,
	}

	tm.tasks.Store(id, task)

	go func() {
		defer cancel()
		
		// Update status to running
		tm.UpdateTask(id, func(t *Task) {
			t.Status = TaskStatusRunning
			t.Message = "Started"
		})

		updateProgress := func(current int, msg string) error {
			// Check if context is cancelled
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			
			tm.UpdateTask(id, func(t *Task) {
				t.Progress = current
				t.Message = msg
			})
			return nil
		}

		// Run the task
		defer func() {
			if r := recover(); r != nil {
				tm.UpdateTask(id, func(t *Task) {
					t.Status = TaskStatusFailed
					t.Error = fmt.Sprintf("Panic: %v", r)
				})
			}
		}()

		err := runFunc(ctx, updateProgress)
		
		tm.UpdateTask(id, func(t *Task) {
			if ctx.Err() != nil {
				// Already handled by StopTask or cancelled
				if t.Status != TaskStatusStopped {
					t.Status = TaskStatusStopped
					t.Message = "Cancelled"
				}
			} else if err != nil {
				t.Status = TaskStatusFailed
				t.Error = err.Error()
			} else {
				t.Status = TaskStatusCompleted
				t.Progress = t.Total
				t.Message = "Completed"
			}
		})
	}()

	return id
}

func (tm *TaskManager) UpdateTask(id string, updateFunc func(*Task)) {
	if val, ok := tm.tasks.Load(id); ok {
		task := val.(*Task)
		task.mu.Lock()
		defer task.mu.Unlock()
		updateFunc(task)
	}
}

func (tm *TaskManager) GetTask(id string) (Task, bool) {
	if val, ok := tm.tasks.Load(id); ok {
		task := val.(*Task)
		task.mu.RLock()
		defer task.mu.RUnlock()
		// Return a copy
		return Task{
			ID:        task.ID,
			Status:    task.Status,
			Progress:  task.Progress,
			Total:     task.Total,
			Message:   task.Message,
			Result:    task.Result,
			Error:     task.Error,
			CreatedAt: task.CreatedAt,
		}, true
	}
	return Task{}, false
}

func (tm *TaskManager) StopTask(id string) {
	if val, ok := tm.tasks.Load(id); ok {
		task := val.(*Task)
		task.mu.Lock()
		defer task.mu.Unlock()
		
		if task.Status == TaskStatusRunning || task.Status == TaskStatusPending {
			task.cancel()
			task.Status = TaskStatusStopped
			task.Message = "Stopped by user"
		}
	}
}
