package ports

import (
	"context"

	"moulaybdl/zindy/task_service/internal/core/domain"

	"github.com/google/uuid"
	// "github.com/google/uuid"
)

// TaskService defines the contract for task business logic
type TaskService interface {
	// Task operations
	CreateTask(ctx context.Context, req *domain.CreateTaskRequest) (*domain.Task, error)
	GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error)
	UpdateTask(ctx context.Context, id uuid.UUID, req *domain.UpdateTaskRequest) (*domain.Task, error)
	DeleteTask(ctx context.Context, id uuid.UUID) error
	
	// // Task queries
	// GetTasksByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Task, error)
	// GetTasksByUser(ctx context.Context, userID uuid.UUID) ([]domain.Task, error)
	
	// // Task assignment and status
	// AssignTaskToMember(ctx context.Context, taskID uuid.UUID, req *domain.AssignTaskRequest) (*domain.TaskData, error)
	// ChangeTaskStatus(ctx context.Context, taskID uuid.UUID, status domain.TaskStatus) error
}
