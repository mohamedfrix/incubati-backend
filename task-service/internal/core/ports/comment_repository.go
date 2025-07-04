package ports

import (
	"context"

	"moulaybdl/zindy/task_service/internal/core/domain"

	"github.com/google/uuid"
)

// CommentRepository defines the contract for comment data operations
type CommentRepository interface {
	// Comment CRUD operations
	Create(ctx context.Context, taskID uuid.UUID, comment *domain.CreateCommentRequest) (*domain.Comment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Comment, error)
	GetByTaskID(ctx context.Context, taskID uuid.UUID) ([]domain.Comment, error)
	Update(ctx context.Context, id uuid.UUID, content string) (*domain.Comment, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
