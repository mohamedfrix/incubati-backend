package domain

import (
	"time"

	"github.com/google/uuid"
)

// Comment represents a comment on a task
type Comment struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Content   string    `json:"content" db:"content"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	TaskID    uuid.UUID `json:"task_id" db:"task_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateCommentRequest represents the request payload for creating a comment
type CreateCommentRequest struct {
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	Content string    `json:"content" validate:"required"`
}

// CommentResponse represents the response structure for comment operations
type CommentResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Data    *Comment `json:"data,omitempty"`
}

// CommentsResponse represents the response structure for multiple comments
type CommentsResponse struct {
	Success bool      `json:"success"`
	Data    []Comment `json:"data"`
}
