package ports

import (
	"context"

	"moulaybdl/zindy/task_service/internal/core/domain"

	"github.com/google/uuid"
)

// CommentRepository defines the contract for task comment data operations
type CommentRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, taskID uuid.UUID, comment *domain.TaskComment) (*domain.TaskComment, error)
	GetByID(ctx context.Context, commentID uuid.UUID) (*domain.TaskComment, error)
	Update(ctx context.Context, commentID uuid.UUID, req *domain.UpdateCommentRequest) (*domain.TaskComment, error)
	Delete(ctx context.Context, commentID uuid.UUID) error
	
	// Comment queries
	GetByTaskID(ctx context.Context, taskID uuid.UUID, pagination *domain.PaginationInfo) ([]domain.TaskComment, *domain.PaginationInfo, error)
	GetCommentCount(ctx context.Context, taskID uuid.UUID) (int, error)
	
	// Validation helpers
	ExistsByID(ctx context.Context, commentID uuid.UUID) (bool, error)
	IsAuthor(ctx context.Context, commentID uuid.UUID, userID uuid.UUID) (bool, error)
}