package domain

import "errors"

// Common errors used across the domain
var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrCommentNotFound   = errors.New("comment not found")
	ErrInvalidTaskStatus = errors.New("invalid task status")
	ErrInvalidPriority   = errors.New("invalid task priority")
	ErrUnauthorized      = errors.New("unauthorized access")
	ErrInvalidUUID       = errors.New("invalid UUID format")
	ErrDuplicateTask     = errors.New("task already exists")
	ErrTaskAssignment    = errors.New("failed to assign task")
)

// BaseResponse represents a basic API response structure
type BaseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ErrorResponse represents an error response structure
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}
