package ports

import (
	"context"

	"moulaybdl/zindy/task_service/internal/core/domain"

	"github.com/google/uuid"
)

// TaskRepository defines the contract for task data operations
type TaskRepository interface {
	// Task CRUD operations
	Create(ctx context.Context, task *domain.Task) (*domain.Task, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error)
	Update(ctx context.Context, id uuid.UUID, task *domain.UpdateTaskRequest) (*domain.Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
	
	// Task queries
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]domain.Task, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Task, error)
	GetByMilestoneID(ctx context.Context, milestoneID uuid.UUID) ([]domain.Task, error)
	GetByStatus(ctx context.Context, status domain.TaskStatus) ([]domain.Task, error)
	
	// Task assignment and status updates
	AssignTask(ctx context.Context, taskID uuid.UUID, req *domain.AssignTaskRequest) (*domain.Task, error)
	UpdateStatus(ctx context.Context, taskID uuid.UUID, status domain.TaskStatus) error
}
