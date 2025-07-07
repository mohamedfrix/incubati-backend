package ports

import (
	"context"

	"moulaybdl/zindy/task_service/internal/core/domain"

	"github.com/google/uuid"
)

// CommentService defines the contract for task comment business logic
type CommentService interface {
	// Comment Management Endpoints
	
	// 9. Add Comment - POST /api/tasks/{task_id}/comments
	AddComment(ctx context.Context, taskID uuid.UUID, req *domain.CreateCommentRequest, userID uuid.UUID) (*domain.CommentResponse, error)
	
	// 10. Get Task Comments - GET /api/tasks/{task_id}/comments
	GetTaskComments(ctx context.Context, taskID uuid.UUID, pagination *domain.PaginationInfo, userID uuid.UUID) (*domain.CommentListResponse, error)
	
	// Additional comment operations
	UpdateComment(ctx context.Context, commentID uuid.UUID, req *domain.UpdateCommentRequest, userID uuid.UUID) (*domain.CommentResponse, error)
	DeleteComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID) (*domain.SuccessResponse, error)
}
