package domain

import (
	"errors"

	"github.com/google/uuid"
)

// Common errors used across the domain
var (
	// Task errors
	ErrTaskNotFound           = errors.New("task not found")
	ErrTaskHasSubtasks        = errors.New("cannot delete task with subtasks")
	ErrInvalidTaskStatus      = errors.New("invalid task status")
	ErrInvalidTaskPriority    = errors.New("invalid task priority")
	ErrSelfReferenceTask      = errors.New("task cannot be parent of itself")
	ErrCircularReference      = errors.New("circular reference detected in task hierarchy")
	ErrDueDateInPast          = errors.New("due date cannot be in the past for new tasks")
	ErrEstimatedHoursInvalid  = errors.New("estimated hours must be greater than 0")
	ErrActualHoursInvalid     = errors.New("actual hours cannot exceed 999 hours")
	
	// Comment errors
	ErrCommentNotFound        = errors.New("comment not found")
	ErrCommentContentEmpty    = errors.New("comment content cannot be empty")
	ErrCommentTooLong         = errors.New("comment content exceeds maximum length")
	
	// Attachment errors
	ErrAttachmentNotFound     = errors.New("attachment not found")
	ErrInvalidMimeType        = errors.New("invalid or unsupported file type")
	ErrFileSizeExceeded       = errors.New("file size exceeds maximum allowed limit")
	ErrInvalidFileName        = errors.New("invalid file name")
	
	// Permission errors
	ErrUnauthorized           = errors.New("unauthorized access")
	ErrInsufficientPermissions = errors.New("insufficient permissions for this operation")
	ErrNotTaskCreator         = errors.New("only task creator can perform this action")
	ErrNotAssignedUser        = errors.New("only assigned user can perform this action")
	
	// Validation errors
	ErrInvalidUUID            = errors.New("invalid UUID format")
	ErrRequiredFieldMissing   = errors.New("required field is missing")
	ErrInvalidPageNumber      = errors.New("invalid page number")
	ErrInvalidPageSize        = errors.New("invalid page size")
	ErrInvalidSearchTerm      = errors.New("invalid search term")
	ErrInvalidSortField       = errors.New("invalid sort field")
	
	// Bulk operation errors
	ErrBulkOperationFailed    = errors.New("bulk operation failed")
	ErrTooManyTasksSelected   = errors.New("too many tasks selected for bulk operation")
	ErrNoTasksSelected        = errors.New("no tasks selected for bulk operation")
	ErrNoUpdateFieldsProvided = errors.New("no update fields provided")
)

// BaseResponse represents a basic API response structure
type BaseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ErrorResponse represents an error response structure
type ErrorResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Error   string      `json:"error,omitempty"`
	Errors  interface{} `json:"errors,omitempty"` // For validation errors
}

// SuccessResponse represents a success response without data
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ValidationError represents a validation error with field-specific messages
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// PermissionContext represents the context for permission checking
type PermissionContext struct {
	UserID           uuid.UUID
	IsAuthenticated  bool
	TaskCreatorID    *uuid.UUID
	TaskAssigneeID   *uuid.UUID
	CommentAuthorID  *uuid.UUID
}

// CanEditTask checks if user can edit a task
func (pc *PermissionContext) CanEditTask() bool {
	if !pc.IsAuthenticated {
		return false
	}
	
	// Creator or assigned user can edit
	if pc.TaskCreatorID != nil && *pc.TaskCreatorID == pc.UserID {
		return true
	}
	
	if pc.TaskAssigneeID != nil && *pc.TaskAssigneeID == pc.UserID {
		return true
	}
	
	return false
}

// CanDeleteTask checks if user can delete a task
func (pc *PermissionContext) CanDeleteTask() bool {
	if !pc.IsAuthenticated {
		return false
	}
	
	// Only creator can delete
	return pc.TaskCreatorID != nil && *pc.TaskCreatorID == pc.UserID
}

// CanAssignTask checks if user can assign/unassign a task
func (pc *PermissionContext) CanAssignTask() bool {
	if !pc.IsAuthenticated {
		return false
	}
	
	// Creator or currently assigned user can assign/unassign
	if pc.TaskCreatorID != nil && *pc.TaskCreatorID == pc.UserID {
		return true
	}
	
	if pc.TaskAssigneeID != nil && *pc.TaskAssigneeID == pc.UserID {
		return true
	}
	
	return false
}

// CanChangeStatus checks if user can change task status
func (pc *PermissionContext) CanChangeStatus() bool {
	return pc.CanEditTask()
}

// CanEditComment checks if user can edit a comment
func (pc *PermissionContext) CanEditComment() bool {
	if !pc.IsAuthenticated {
		return false
	}
	
	// Only comment author can edit
	return pc.CommentAuthorID != nil && *pc.CommentAuthorID == pc.UserID
}
