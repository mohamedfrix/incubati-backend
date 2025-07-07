package ports

import (
	"context"

	"moulaybdl/zindy/task_service/internal/core/domain"

	"github.com/google/uuid"
)

// TaskRepository defines the contract for task data operations
type TaskRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, task *domain.Task) (*domain.Task, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error)
	Update(ctx context.Context, id uuid.UUID, req *domain.UpdateTaskRequest) (*domain.Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Task queries for different endpoints
	GetTasksByProject(ctx context.Context, projectID uuid.UUID, filter *domain.TaskFilterRequest) ([]domain.Task, *domain.PaginationInfo, error)
	GetTasksByUser(ctx context.Context, userID uuid.UUID, filter *domain.TaskFilterRequest) ([]domain.Task, *domain.PaginationInfo, error)
	
	// Task assignment operations
	AssignTask(ctx context.Context, taskID uuid.UUID, req *domain.AssignTaskRequest) (*domain.Task, error)
	UnassignTask(ctx context.Context, taskID uuid.UUID) (*domain.Task, error)
	
	// Status management
	ChangeStatus(ctx context.Context, taskID uuid.UUID, req *domain.ChangeStatusRequest) (*domain.Task, error)
	
	// Bulk operations
	BulkUpdate(ctx context.Context, req *domain.BulkUpdateTasksRequest) ([]domain.Task, error)
	
	// Statistics and analytics
	GetTaskStatistics(ctx context.Context, taskID uuid.UUID) (*domain.TaskStatistics, error)
	GetTaskSummary(ctx context.Context, projectID *uuid.UUID, userID *uuid.UUID) (*domain.TaskSummary, error)
	
	// Subtask operations
	GetSubtasks(ctx context.Context, parentTaskID uuid.UUID, includeSubtasks bool) ([]domain.Task, error)
	
	// Validation helpers
	HasSubtasks(ctx context.Context, taskID uuid.UUID) (bool, error)
	ValidateParentTask(ctx context.Context, taskID uuid.UUID, parentTaskID *uuid.UUID) error
}
