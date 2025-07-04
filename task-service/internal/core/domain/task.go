package domain

import (
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents the possible states of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusActive    TaskStatus = "active"
	TaskStatusDone      TaskStatus = "done"
	TaskStatusOnHold    TaskStatus = "on_hold"
	TaskStatusTesting   TaskStatus = "testing"
	TaskStatusPostponed TaskStatus = "postponed"
)

// TaskPriority represents the priority levels of a task
type TaskPriority string

const (
	TaskPriorityLow       TaskPriority = "low"
	TaskPriorityMedium    TaskPriority = "medium"
	TaskPriorityHigh      TaskPriority = "high"
	TaskPriorityUrgent    TaskPriority = "urgent"
	TaskPriorityEmergency TaskPriority = "emergency"
)

// Task represents a task in the system
type Task struct {
	ID          uuid.UUID    `json:"id" db:"id"`
	Title       string       `json:"title" db:"title"`
	Description *string      `json:"description" db:"description"`
	Status      TaskStatus   `json:"status" db:"status"`
	Priority    TaskPriority `json:"priority" db:"priority"`
	DueDate     *time.Time   `json:"due_date" db:"due_date"`
	CompletedAt *time.Time   `json:"completed_at" db:"completed_at"`
	AssigneeID  uuid.UUID    `json:"assignee_id" db:"assignee_id"`
	ProjectID   uuid.UUID    `json:"project_id" db:"project_id"`
	MilestoneID uuid.UUID    `json:"milestone_id" db:"milestone_id"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	Comments    []Comment    `json:"comments,omitempty"`
}

// CreateTaskRequest represents the request payload for creating a task
type CreateTaskRequest struct {
	Title       string        `json:"title" validate:"required,max=255"`
	Description *string       `json:"description"`
	Status      *TaskStatus   `json:"status"`
	Priority    *TaskPriority `json:"priority"`
	DueDate     *time.Time    `json:"due_date"`
	CompletedAt *time.Time    `json:"completed_at"`
	AssigneeID  *uuid.UUID    `json:"assignee_id"`
	ProjectID   uuid.UUID     `json:"project_id" validate:"required"`
	MilestoneID uuid.UUID     `json:"milestone_id" validate:"required"`
}

// UpdateTaskRequest represents the request payload for updating a task
type UpdateTaskRequest struct {
	Title       *string       `json:"title" validate:"omitempty,max=255"`
	Description *string       `json:"description"`
	Status      *TaskStatus   `json:"status"`
	Priority    *TaskPriority `json:"priority"`
	DueDate     *time.Time    `json:"due_date"`
	CompletedAt *time.Time    `json:"completed_at"`
	AssigneeID  *uuid.UUID    `json:"assignee_id"`
	ProjectID   *uuid.UUID    `json:"project_id"`
	MilestoneID *uuid.UUID    `json:"milestone_id"`
}

// AssignTaskRequest represents the request payload for assigning a task to a member
type AssignTaskRequest struct {
	UserID      uuid.UUID     `json:"user_id" validate:"required"`
	DueDate     *time.Time    `json:"due_date"`
	CompletedAt *time.Time    `json:"completed_date"`
	Status      *TaskStatus   `json:"status"`
	Priority    *TaskPriority `json:"priority"`
}

// UpdateTaskStatusRequest represents the request payload for updating task status
type UpdateTaskStatusRequest struct {
	Status TaskStatus `json:"status" validate:"required"`
}

// TaskResponse represents the response structure for task operations
type TaskResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Task    *Task  `json:"task,omitempty"`
}

// TasksResponse represents the response structure for multiple tasks
type TasksResponse struct {
	Success bool   `json:"success"`
	Data    []Task `json:"data"`
}

// AssignTaskResponse represents the response structure for task assignment
type AssignTaskResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Data    *TaskData `json:"data,omitempty"`
}

// TaskData represents the data structure for task assignment response
type TaskData struct {
	TaskID uuid.UUID `json:"task_id"`
	UserID uuid.UUID `json:"user_id"`
	Task   *Task     `json:"task"`
}

// IsValidStatus checks if the task status is valid
func (ts TaskStatus) IsValid() bool {
	switch ts {
	case TaskStatusPending, TaskStatusActive, TaskStatusDone, TaskStatusOnHold, TaskStatusTesting, TaskStatusPostponed:
		return true
	}
	return false
}

// IsValidPriority checks if the task priority is valid
func (tp TaskPriority) IsValid() bool {
	switch tp {
	case TaskPriorityLow, TaskPriorityMedium, TaskPriorityHigh, TaskPriorityUrgent, TaskPriorityEmergency:
		return true
	}
	return false
}

// SetDefaults sets default values for task creation
func (t *Task) SetDefaults() {
	if t.Status == "" {
		t.Status = TaskStatusPending
	}
	if t.Priority == "" {
		t.Priority = TaskPriorityLow
	}
	if t.AssigneeID == uuid.Nil {
		t.AssigneeID = uuid.MustParse("00000000-0000-0000-0000-000000000000")
	}
}
