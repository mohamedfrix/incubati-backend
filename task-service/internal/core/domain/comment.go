package domain

import (
	"time"

	"github.com/google/uuid"
)

// TaskComment represents a comment on a task
type TaskComment struct {
	ID             uuid.UUID `json:"id" db:"id"`
	TaskID         uuid.UUID `json:"task_id" db:"task_id"`
	AuthorUserID   uuid.UUID `json:"author_user_id" db:"author_user_id"`
	Content        string    `json:"content" db:"content"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// CreateCommentRequest represents the request payload for creating a comment
type CreateCommentRequest struct {
	AuthorUserID *uuid.UUID `json:"author_user_id"` // Auto-set if not provided
	Content      string     `json:"content" validate:"required,max=5000"`
}

// UpdateCommentRequest represents the request payload for updating a comment
type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required,max=5000"`
}

// CommentResponse represents the response structure for comment operations
type CommentResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Data    *TaskComment `json:"data,omitempty"`
}

// CommentListResponse represents the response structure for comment lists
type CommentListResponse struct {
	Success bool             `json:"success"`
	Data    CommentListData  `json:"data"`
}

// CommentListData represents the data structure for comment list responses
type CommentListData struct {
	Comments   []TaskComment  `json:"comments"`
	Pagination PaginationInfo `json:"pagination"`
}

// CommentWithDisplayData represents a comment with additional display information
type CommentWithDisplayData struct {
	TaskComment
	CreatedAtDisplay string `json:"created_at_display"`
	IsEdited         bool   `json:"is_edited"`
}

// IsEdited checks if the comment has been edited
func (c *TaskComment) IsEdited() bool {
	return !c.CreatedAt.Equal(c.UpdatedAt)
}

// GetDisplayData returns comment with additional display information
func (c *TaskComment) GetDisplayData() CommentWithDisplayData {
	return CommentWithDisplayData{
		TaskComment:      *c,
		CreatedAtDisplay: c.CreatedAt.Format("2006-01-02 15:04:05"),
		IsEdited:         c.IsEdited(),
	}
}
