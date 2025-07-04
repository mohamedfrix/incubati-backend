package ports

import (
	"context"

	"moulaybdl/zindy/task_service/internal/core/domain"

	"github.com/google/uuid"
)

// CommentService defines the contract for comment business logic
type CommentService interface {
	AddComment(ctx context.Context, taskID uuid.UUID, req *domain.CreateCommentRequest) (*domain.Comment, error)
	GetCommentsByTask(ctx context.Context, taskID uuid.UUID) ([]domain.Comment, error)
	UpdateComment(ctx context.Context, commentID uuid.UUID, content string) (*domain.Comment, error)
	DeleteComment(ctx context.Context, commentID uuid.UUID) error
}
