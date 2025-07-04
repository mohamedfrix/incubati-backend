package domain

import (
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents the possible states of a task
type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusReview     TaskStatus = "review"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// TaskPriority represents the priority levels of a task
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

// Task represents a task in the system
type Task struct {
	ID                 uuid.UUID     `json:"id" db:"id"`
	ProjectID          uuid.UUID     `json:"project_id" db:"project_id"`
	MilestoneID        uuid.UUID     `json:"milestone_id" db:"milestone_id"`
	AssignedToUserID   *uuid.UUID    `json:"assigned_to_user_id" db:"assigned_to_user_id"`
	CreatedByUserID    uuid.UUID     `json:"created_by_user_id" db:"created_by_user_id"`
	Title              string        `json:"title" db:"title"`
	Description        *string       `json:"description" db:"description"`
	Status             TaskStatus    `json:"status" db:"status"`
	Priority           TaskPriority  `json:"priority" db:"priority"`
	CreatedAt          time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at" db:"updated_at"`
	DueDate            *time.Time    `json:"due_date" db:"due_date"`
	CompletedAt        *time.Time    `json:"completed_at" db:"completed_at"`
	EstimatedHours     *float64      `json:"estimated_hours" db:"estimated_hours"`
	ActualHours        *float64      `json:"actual_hours" db:"actual_hours"`
	ParentTaskID       *uuid.UUID    `json:"parent_task_id" db:"parent_task_id"`
	
	// Related data (loaded separately)
	Comments           []TaskComment    `json:"comments,omitempty"`
	Attachments        []TaskAttachment `json:"attachments,omitempty"`
	Subtasks           []Task           `json:"subtasks,omitempty"`
}

// CreateTaskRequest represents the request payload for creating a task
type CreateTaskRequest struct {
	ProjectID          uuid.UUID     `json:"project_id" validate:"required"`
	MilestoneID        uuid.UUID     `json:"milestone_id" validate:"required"`
	AssignedToUserID   *uuid.UUID    `json:"assigned_to_user_id"`
	CreatedByUserID    *uuid.UUID    `json:"created_by_user_id"` // Auto-set if not provided
	Title              string        `json:"title" validate:"required,max=255"`
	Description        *string       `json:"description"`
	Status             *TaskStatus   `json:"status"`
	Priority           *TaskPriority `json:"priority"`
	DueDate            *time.Time    `json:"due_date"`
	EstimatedHours     *float64      `json:"estimated_hours" validate:"omitempty,gt=0"`
	ActualHours        *float64      `json:"actual_hours" validate:"omitempty,gte=0,lte=999"`
	ParentTaskID       *uuid.UUID    `json:"parent_task_id"`
}

// UpdateTaskRequest represents the request payload for updating a task
type UpdateTaskRequest struct {
	Title              *string       `json:"title" validate:"omitempty,max=255"`
	Description        *string       `json:"description"`
	Status             *TaskStatus   `json:"status"`
	Priority           *TaskPriority `json:"priority"`
	DueDate            *time.Time    `json:"due_date"`
	EstimatedHours     *float64      `json:"estimated_hours" validate:"omitempty,gt=0"`
	ActualHours        *float64      `json:"actual_hours" validate:"omitempty,gte=0,lte=999"`
	AssignedToUserID   *uuid.UUID    `json:"assigned_to_user_id"`
	ProjectID          *uuid.UUID    `json:"project_id"`
	MilestoneID        *uuid.UUID    `json:"milestone_id"`
	ParentTaskID       *uuid.UUID    `json:"parent_task_id"`
}

// AssignTaskRequest represents the request payload for assigning a task
type AssignTaskRequest struct {
	AssignedToUserID *uuid.UUID `json:"assigned_to_user_id" validate:"required"`
}

// ChangeStatusRequest represents the request payload for changing task status
type ChangeStatusRequest struct {
	Status TaskStatus `json:"status" validate:"required"`
}

// BulkUpdateTasksRequest represents the request payload for bulk updating tasks
type BulkUpdateTasksRequest struct {
	TaskIDs            []uuid.UUID   `json:"task_ids" validate:"required,min=1,max=100"`
	Status             *TaskStatus   `json:"status"`
	Priority           *TaskPriority `json:"priority"`
	AssignedToUserID   *uuid.UUID    `json:"assigned_to_user_id"`
	DueDate            *time.Time    `json:"due_date"`
}

// TaskFilterRequest represents query parameters for filtering tasks
type TaskFilterRequest struct {
	Status             *TaskStatus `form:"status"`
	Priority           *TaskPriority `form:"priority"`
	AssignedToUserID   *uuid.UUID  `form:"assigned_to_user_id"`
	ProjectID          *uuid.UUID  `form:"project_id"`
	Search             *string     `form:"search"`
	IncludeSubtasks    *bool       `form:"include_subtasks"`
	OverdueOnly        *bool       `form:"overdue_only"`
	DueSoon            *int        `form:"due_soon"` // Days
	OrderBy            *string     `form:"order_by"`
	Page               *int        `form:"page"`
	PageSize           *int        `form:"page_size"`
}

// TaskResponse represents the response structure for single task operations
type TaskResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    *Task  `json:"data,omitempty"`
}

// TaskListResponse represents the response structure for task lists
type TaskListResponse struct {
	Success bool        `json:"success"`
	Data    TaskListData `json:"data"`
}

// TaskListData represents the data structure for task list responses
type TaskListData struct {
	Tasks      []Task         `json:"tasks"`
	Pagination PaginationInfo `json:"pagination"`
	Summary    *TaskSummary   `json:"summary,omitempty"`
}

// TaskSummary represents summary statistics for task lists
type TaskSummary struct {
	TotalTasks int                    `json:"total_tasks"`
	ByStatus   map[TaskStatus]int     `json:"by_status"`
	ByPriority map[TaskPriority]int   `json:"by_priority"`
}

// TaskStatistics represents detailed statistics for a task
type TaskStatistics struct {
	Task           Task                `json:"task"`
	SubtaskStats   SubtaskStatistics   `json:"subtask_stats"`
	CommentStats   CommentStatistics   `json:"comment_stats"`
	AttachmentStats AttachmentStatistics `json:"attachment_stats"`
	TimeTracking   TimeTrackingInfo    `json:"time_tracking"`
}

// SubtaskStatistics represents statistics about subtasks
type SubtaskStatistics struct {
	Total       int                `json:"total"`
	Completed   int                `json:"completed"`
	InProgress  int                `json:"in_progress"`
	ByStatus    map[TaskStatus]int `json:"by_status"`
	Completion  float64            `json:"completion_percentage"`
}

// CommentStatistics represents statistics about task comments
type CommentStatistics struct {
	Total            int               `json:"total"`
	UniqueAuthors    int               `json:"unique_authors"`
	LastCommentAt    *time.Time        `json:"last_comment_at"`
	AuthorActivity   map[string]int    `json:"author_activity"`
}

// AttachmentStatistics represents statistics about task attachments
type AttachmentStatistics struct {
	Total        int     `json:"total"`
	TotalSize    int64   `json:"total_size_bytes"`
	TotalSizeMB  float64 `json:"total_size_mb"`
	ByType       map[string]int `json:"by_type"`
	LastUpload   *time.Time `json:"last_upload_at"`
}

// TimeTrackingInfo represents time tracking information
type TimeTrackingInfo struct {
	EstimatedHours   *float64 `json:"estimated_hours"`
	ActualHours      *float64 `json:"actual_hours"`
	RemainingHours   *float64 `json:"remaining_hours"`
	IsOverEstimate   bool     `json:"is_over_estimate"`
	CompletionRatio  *float64 `json:"completion_ratio"`
}

// PaginationInfo represents pagination information
type PaginationInfo struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

// IsValidStatus checks if the task status is valid
func (ts TaskStatus) IsValid() bool {
	switch ts {
	case TaskStatusTodo, TaskStatusInProgress, TaskStatusReview, TaskStatusDone, TaskStatusCancelled:
		return true
	}
	return false
}

// IsValidPriority checks if the task priority is valid
func (tp TaskPriority) IsValid() bool {
	switch tp {
	case TaskPriorityLow, TaskPriorityMedium, TaskPriorityHigh, TaskPriorityUrgent:
		return true
	}
	return false
}

// SetDefaults sets default values for task creation
func (t *Task) SetDefaults() {
	if t.Status == "" {
		t.Status = TaskStatusTodo
	}
	if t.Priority == "" {
		t.Priority = TaskPriorityMedium
	}
}

// IsOverdue checks if the task is overdue
func (t *Task) IsOverdue() bool {
	if t.DueDate == nil || t.Status == TaskStatusDone {
		return false
	}
	return t.DueDate.Before(time.Now())
}

// CompletionPercentage calculates completion percentage for tasks with subtasks
func (t *Task) CompletionPercentage() float64 {
	if len(t.Subtasks) == 0 {
		if t.Status == TaskStatusDone {
			return 100.0
		}
		return 0.0
	}
	
	completed := 0
	for _, subtask := range t.Subtasks {
		if subtask.Status == TaskStatusDone {
			completed++
		}
	}
	
	return float64(completed) / float64(len(t.Subtasks)) * 100.0
}

// DaysUntilDue calculates days until due date
func (t *Task) DaysUntilDue() *int {
	if t.DueDate == nil {
		return nil
	}
	
	days := int(time.Until(*t.DueDate).Hours() / 24)
	return &days
}
