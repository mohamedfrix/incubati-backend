package services

import (
	"context"
	"strings"

	"moulaybdl/zindy/task_service/internal/core/domain"
	"moulaybdl/zindy/task_service/internal/core/ports"

	"github.com/google/uuid"
)

type CommentService struct {
	commentRepo ports.CommentRepository
	taskRepo    ports.TaskRepository
}

func NewCommentService(commentRepo ports.CommentRepository, taskRepo ports.TaskRepository) ports.CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		taskRepo:    taskRepo,
	}
}

// AddComment implements ports.CommentService
func (s *CommentService) AddComment(ctx context.Context, taskID uuid.UUID, req *domain.CreateCommentRequest, userID uuid.UUID) (*domain.CommentResponse, error) {
	// Validate request
	if err := s.validateCreateCommentRequest(req); err != nil {
		return &domain.CommentResponse{
			Success: false,
			Message: "Validation failed",
		}, err
	}

	// Check if task exists and user has permission to comment
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return &domain.CommentResponse{
			Success: false,
			Message: "Task not found",
		}, err
	}

	// Check permissions - for comments, we allow any authenticated user to comment
	// TODO: Add more sophisticated project/team-based permission checks
	permCtx := &domain.PermissionContext{
		UserID:          userID,
		IsAuthenticated: true,
		TaskCreatorID:   &task.CreatedByUserID,
		TaskAssigneeID:  task.AssignedToUserID,
	}

	if !s.canCommentOnTask(permCtx) {
		return &domain.CommentResponse{
			Success: false,
			Message: "Unauthorized to comment on this task",
		}, domain.ErrUnauthorized
	}

	// Set user as author if not provided
	if req.AuthorUserID == nil {
		req.AuthorUserID = &userID
	}

	// Create comment
	comment := &domain.TaskComment{
		AuthorUserID: *req.AuthorUserID,
		Content:      strings.TrimSpace(req.Content),
	}

	createdComment, err := s.commentRepo.Create(ctx, taskID, comment)
	if err != nil {
		return &domain.CommentResponse{
			Success: false,
			Message: "Failed to create comment",
		}, err
	}

	return &domain.CommentResponse{
		Success: true,
		Message: "Comment added successfully",
		Data:    createdComment,
	}, nil
}

// GetTaskComments implements ports.CommentService
func (s *CommentService) GetTaskComments(ctx context.Context, taskID uuid.UUID, pagination *domain.PaginationInfo, userID uuid.UUID) (*domain.CommentListResponse, error) {
	// Check if task exists and user has permission to view comments
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return &domain.CommentListResponse{
			Success: false,
			Data: domain.CommentListData{
				Comments:   []domain.TaskComment{},
				Pagination: domain.PaginationInfo{},
			},
		}, err
	}

	// Check permissions
	permCtx := &domain.PermissionContext{
		UserID:          userID,
		IsAuthenticated: true,
		TaskCreatorID:   &task.CreatedByUserID,
		TaskAssigneeID:  task.AssignedToUserID,
	}

	if !s.canViewTask(permCtx) {
		return &domain.CommentListResponse{
			Success: false,
			Data: domain.CommentListData{
				Comments:   []domain.TaskComment{},
				Pagination: domain.PaginationInfo{},
			},
		}, domain.ErrUnauthorized
	}

	// Get comments
	comments, paginationInfo, err := s.commentRepo.GetByTaskID(ctx, taskID, pagination)
	if err != nil {
		return &domain.CommentListResponse{
			Success: false,
			Data: domain.CommentListData{
				Comments:   []domain.TaskComment{},
				Pagination: domain.PaginationInfo{},
			},
		}, err
	}

	return &domain.CommentListResponse{
		Success: true,
		Data: domain.CommentListData{
			Comments:   comments,
			Pagination: *paginationInfo,
		},
	}, nil
}

// UpdateComment implements ports.CommentService
func (s *CommentService) UpdateComment(ctx context.Context, commentID uuid.UUID, req *domain.UpdateCommentRequest, userID uuid.UUID) (*domain.CommentResponse, error) {
	// Validate request
	if err := s.validateUpdateCommentRequest(req); err != nil {
		return &domain.CommentResponse{
			Success: false,
			Message: "Validation failed",
		}, err
	}

	// Check if comment exists
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return &domain.CommentResponse{
			Success: false,
			Message: "Comment not found",
		}, err
	}

	// Check permissions - only author can edit
	permCtx := &domain.PermissionContext{
		UserID:          userID,
		IsAuthenticated: true,
		CommentAuthorID: &comment.AuthorUserID,
	}

	if !permCtx.CanEditComment() {
		return &domain.CommentResponse{
			Success: false,
			Message: "Only comment author can edit this comment",
		}, domain.ErrInsufficientPermissions
	}

	// Trim content
	req.Content = strings.TrimSpace(req.Content)

	// Update comment
	updatedComment, err := s.commentRepo.Update(ctx, commentID, req)
	if err != nil {
		return &domain.CommentResponse{
			Success: false,
			Message: "Failed to update comment",
		}, err
	}

	return &domain.CommentResponse{
		Success: true,
		Message: "Comment updated successfully",
		Data:    updatedComment,
	}, nil
}

// DeleteComment implements ports.CommentService
func (s *CommentService) DeleteComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID) (*domain.SuccessResponse, error) {
	// Check if comment exists
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return &domain.SuccessResponse{
			Success: false,
			Message: "Comment not found",
		}, err
	}

	// Check permissions - only author can delete
	permCtx := &domain.PermissionContext{
		UserID:          userID,
		IsAuthenticated: true,
		CommentAuthorID: &comment.AuthorUserID,
	}

	if !permCtx.CanEditComment() {
		return &domain.SuccessResponse{
			Success: false,
			Message: "Only comment author can delete this comment",
		}, domain.ErrInsufficientPermissions
	}

	// Delete comment
	if err := s.commentRepo.Delete(ctx, commentID); err != nil {
		return &domain.SuccessResponse{
			Success: false,
			Message: "Failed to delete comment",
		}, err
	}

	return &domain.SuccessResponse{
		Success: true,
		Message: "Comment deleted successfully",
	}, nil
}

// Helper methods

func (s *CommentService) validateCreateCommentRequest(req *domain.CreateCommentRequest) error {
	if strings.TrimSpace(req.Content) == "" {
		return domain.ErrCommentContentEmpty
	}

	if len(req.Content) > 5000 {
		return domain.ErrCommentTooLong
	}

	return nil
}

func (s *CommentService) validateUpdateCommentRequest(req *domain.UpdateCommentRequest) error {
	if strings.TrimSpace(req.Content) == "" {
		return domain.ErrCommentContentEmpty
	}

	if len(req.Content) > 5000 {
		return domain.ErrCommentTooLong
	}

	return nil
}

func (s *CommentService) canCommentOnTask(permCtx *domain.PermissionContext) bool {
	if !permCtx.IsAuthenticated {
		return false
	}
	
	// For now, allow any authenticated user to comment
	// TODO: Implement proper project/team-based permissions
	return true
}

func (s *CommentService) canViewTask(permCtx *domain.PermissionContext) bool {
	if !permCtx.IsAuthenticated {
		return false
	}
	
	// For now, allow any authenticated user to view tasks
	// TODO: Implement proper project/team-based permissions
	return true
}
