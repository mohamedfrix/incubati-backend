package ports

import (
	"context"

	"moulaybdl/zindy/task_service/internal/core/domain"

	"github.com/google/uuid"
)

// TaskService defines the contract for task business logic
type TaskService interface {
	// Task Management Endpoints
	
	// 1. Create Task - POST /api/tasks
	CreateTask(ctx context.Context, req *domain.CreateTaskRequest, userID uuid.UUID) (*domain.TaskResponse, error)
	
	// 2. Get Task by ID - GET /api/tasks/{task_id}
	GetTaskByID(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (*domain.TaskResponse, error)
	
	// 3. Update Task - PUT/PATCH /api/tasks/{task_id}
	UpdateTask(ctx context.Context, taskID uuid.UUID, req *domain.UpdateTaskRequest, userID uuid.UUID) (*domain.TaskResponse, error)
	
	// 4. Delete Task - DELETE /api/tasks/{task_id}
	DeleteTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (*domain.SuccessResponse, error)
	
	// 5. Assign/Unassign Task - POST/DELETE /api/tasks/{task_id}/assign
	AssignTask(ctx context.Context, taskID uuid.UUID, req *domain.AssignTaskRequest, userID uuid.UUID) (*domain.TaskResponse, error)
	UnassignTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (*domain.TaskResponse, error)
	
	// 6. Change Task Status - POST /api/tasks/{task_id}/status
	ChangeTaskStatus(ctx context.Context, taskID uuid.UUID, req *domain.ChangeStatusRequest, userID uuid.UUID) (*domain.SuccessResponse, error)
	
	// Task Query Endpoints
	
	// 7. Get Tasks by Project - GET /api/projects/{project_id}/tasks
	GetTasksByProject(ctx context.Context, projectID uuid.UUID, filter *domain.TaskFilterRequest, userID uuid.UUID) (*domain.TaskListResponse, error)
	
	// 8. Get Tasks by User - GET /api/users/{user_id}/tasks
	GetTasksByUser(ctx context.Context, userID uuid.UUID, filter *domain.TaskFilterRequest, requestUserID uuid.UUID) (*domain.TaskListResponse, error)
	
	// Bulk Operations
	
	// 11. Bulk Update Tasks - POST /api/tasks/bulk-update
	BulkUpdateTasks(ctx context.Context, req *domain.BulkUpdateTasksRequest, userID uuid.UUID) (*domain.TaskListResponse, error)
	
	// Analytics
	
	// 12. Get Task Statistics - GET /api/tasks/{task_id}/statistics
	GetTaskStatistics(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (*domain.TaskStatistics, error)
}
