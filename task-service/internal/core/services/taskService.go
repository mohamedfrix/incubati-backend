package services

import (
	"context"
	"fmt"
	"time"

	"moulaybdl/zindy/task_service/internal/core/domain"
	"moulaybdl/zindy/task_service/internal/core/ports"

	"github.com/google/uuid"
)

type TaskService struct {
	taskRepo ports.TaskRepository
}

func NewTaskService(taskRepo ports.TaskRepository) ports.TaskService {
	return &TaskService{
		taskRepo: taskRepo,
	}
}

// CreateTask implements ports.TaskService
func (s *TaskService) CreateTask(ctx context.Context, req *domain.CreateTaskRequest, userID uuid.UUID) (*domain.TaskResponse, error) {
	// Validate request
	if err := s.validateCreateTaskRequest(req); err != nil {
		return &domain.TaskResponse{
			Success: false,
			Message: "Validation failed",
		}, err
	}

	// Set user as creator if not provided
	if req.CreatedByUserID == nil {
		req.CreatedByUserID = &userID
	}

	// Convert request to task domain object
	task := &domain.Task{
		ProjectID:        req.ProjectID,
		MilestoneID:      req.MilestoneID,
		AssignedToUserID: req.AssignedToUserID,
		CreatedByUserID:  *req.CreatedByUserID,
		Title:            req.Title,
		Description:      req.Description,
		DueDate:          req.DueDate,
		EstimatedHours:   req.EstimatedHours,
		ActualHours:      req.ActualHours,
		ParentTaskID:     req.ParentTaskID,
	}

	// Set status and priority with defaults
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	task.SetDefaults()


	// Validate parent task if provided
	if task.ParentTaskID != nil {
		if err := s.taskRepo.ValidateParentTask(ctx, task.ID, task.ParentTaskID); err != nil {
			return &domain.TaskResponse{
				Success: false,
				Message: "Invalid parent task",
			}, err
		}
	}

	// Create task
	createdTask, err := s.taskRepo.Create(ctx, task)
	if err != nil {
		return &domain.TaskResponse{
			Success: false,
			Message: "Failed to create task",
		}, err
	}

	return &domain.TaskResponse{
		Success: true,
		Message: "Task created successfully",
		Data:    createdTask,
	}, nil
}

// GetTaskByID implements ports.TaskService
func (s *TaskService) GetTaskByID(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (*domain.TaskResponse, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return &domain.TaskResponse{
			Success: false,
			Message: "Task not found",
		}, err
	}

	// Check permissions
	permCtx := &domain.PermissionContext{
		UserID:          userID,
		IsAuthenticated: true,
		TaskCreatorID:   &task.CreatedByUserID,
		TaskAssigneeID:  task.AssignedToUserID,
	}

	// For viewing tasks, we allow broader access (creator, assignee, or team member)
	if !s.canViewTask(permCtx) {
		return &domain.TaskResponse{
			Success: false,
			Message: "Unauthorized access",
		}, domain.ErrUnauthorized
	}

	return &domain.TaskResponse{
		Success: true,
		Message: "Task retrieved successfully",
		Data:    task,
	}, nil
}

// UpdateTask implements ports.TaskService
func (s *TaskService) UpdateTask(ctx context.Context, taskID uuid.UUID, req *domain.UpdateTaskRequest, userID uuid.UUID) (*domain.TaskResponse, error) {
	// Get existing task for permission check
	existingTask, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return &domain.TaskResponse{
			Success: false,
			Message: "Task not found",
		}, err
	}

	// Check permissions
	permCtx := &domain.PermissionContext{
		UserID:          userID,
		IsAuthenticated: true,
		TaskCreatorID:   &existingTask.CreatedByUserID,
		TaskAssigneeID:  existingTask.AssignedToUserID,
	}

	if !permCtx.CanEditTask() {
		return &domain.TaskResponse{
			Success: false,
			Message: "Insufficient permissions",
		}, domain.ErrInsufficientPermissions
	}

	// Validate update request
	if err := s.validateUpdateTaskRequest(req); err != nil {
		return &domain.TaskResponse{
			Success: false,
			Message: "Validation failed",
		}, err
	}

	// Validate parent task change if provided
	if req.ParentTaskID != nil {
		if err := s.taskRepo.ValidateParentTask(ctx, taskID, req.ParentTaskID); err != nil {
			return &domain.TaskResponse{
				Success: false,
				Message: "Invalid parent task",
			}, err
		}
	}

	// Update task
	updatedTask, err := s.taskRepo.Update(ctx, taskID, req)
	if err != nil {
		return &domain.TaskResponse{
			Success: false,
			Message: "Failed to update task",
		}, err
	}

	return &domain.TaskResponse{
		Success: true,
		Message: "Task updated successfully",
		Data:    updatedTask,
	}, nil
}

// DeleteTask implements ports.TaskService
func (s *TaskService) DeleteTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (*domain.SuccessResponse, error) {
	// Get existing task for permission check
	existingTask, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return &domain.SuccessResponse{
			Success: false,
			Message: "Task not found",
		}, err
	}

	// Check permissions
	permCtx := &domain.PermissionContext{
		UserID:          userID,
		IsAuthenticated: true,
		TaskCreatorID:   &existingTask.CreatedByUserID,
		TaskAssigneeID:  existingTask.AssignedToUserID,
	}

	if !permCtx.CanDeleteTask() {
		return &domain.SuccessResponse{
			Success: false,
			Message: "Only the task creator can delete this task",
		}, domain.ErrNotTaskCreator
	}

	// Check if task has subtasks
	hasSubtasks, err := s.taskRepo.HasSubtasks(ctx, taskID)
	if err != nil {
		return &domain.SuccessResponse{
			Success: false,
			Message: "Failed to check subtasks",
		}, err
	}

	if hasSubtasks {
		return &domain.SuccessResponse{
			Success: false,
			Message: "Cannot delete task with subtasks",
		}, domain.ErrTaskHasSubtasks
	}

	// Delete task
	if err := s.taskRepo.Delete(ctx, taskID); err != nil {
		return &domain.SuccessResponse{
			Success: false,
			Message: "Failed to delete task",
		}, err
	}

	return &domain.SuccessResponse{
		Success: true,
		Message: "Task deleted successfully",
	}, nil
}

// AssignTask implements ports.TaskService
func (s *TaskService) AssignTask(ctx context.Context, taskID uuid.UUID, req *domain.AssignTaskRequest, userID uuid.UUID) (*domain.TaskResponse, error) {
	// Get existing task for permission check
	existingTask, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return &domain.TaskResponse{
			Success: false,
			Message: "Task not found",
		}, err
	}

	// Check permissions
	permCtx := &domain.PermissionContext{
		UserID:          userID,
		IsAuthenticated: true,
		TaskCreatorID:   &existingTask.CreatedByUserID,
		TaskAssigneeID:  existingTask.AssignedToUserID,
	}

	if !permCtx.CanAssignTask() {
		return &domain.TaskResponse{
			Success: false,
			Message: "Insufficient permissions to assign task",
		}, domain.ErrInsufficientPermissions
	}

	// Assign task
	updatedTask, err := s.taskRepo.AssignTask(ctx, taskID, req)
	if err != nil {
		return &domain.TaskResponse{
			Success: false,
			Message: "Failed to assign task",
		}, err
	}

	return &domain.TaskResponse{
		Success: true,
		Message: "Task assigned successfully",
		Data:    updatedTask,
	}, nil
}

// UnassignTask implements ports.TaskService
func (s *TaskService) UnassignTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (*domain.TaskResponse, error) {
	// Get existing task for permission check
	existingTask, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return &domain.TaskResponse{
			Success: false,
			Message: "Task not found",
		}, err
	}

	// Check permissions
	permCtx := &domain.PermissionContext{
		UserID:          userID,
		IsAuthenticated: true,
		TaskCreatorID:   &existingTask.CreatedByUserID,
		TaskAssigneeID:  existingTask.AssignedToUserID,
	}

	if !permCtx.CanAssignTask() {
		return &domain.TaskResponse{
			Success: false,
			Message: "Insufficient permissions to unassign task",
		}, domain.ErrInsufficientPermissions
	}

	// Unassign task
	updatedTask, err := s.taskRepo.UnassignTask(ctx, taskID)
	if err != nil {
		return &domain.TaskResponse{
			Success: false,
			Message: "Failed to unassign task",
		}, err
	}

	return &domain.TaskResponse{
		Success: true,
		Message: "Task unassigned successfully",
		Data:    updatedTask,
	}, nil
}

// ChangeTaskStatus implements ports.TaskService
func (s *TaskService) ChangeTaskStatus(ctx context.Context, taskID uuid.UUID, req *domain.ChangeStatusRequest, userID uuid.UUID) (*domain.SuccessResponse, error) {
	// Get existing task for permission check
	existingTask, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return &domain.SuccessResponse{
			Success: false,
			Message: "Task not found",
		}, err
	}

	// Check permissions
	permCtx := &domain.PermissionContext{
		UserID:          userID,
		IsAuthenticated: true,
		TaskCreatorID:   &existingTask.CreatedByUserID,
		TaskAssigneeID:  existingTask.AssignedToUserID,
	}

	if !permCtx.CanChangeStatus() {
		return &domain.SuccessResponse{
			Success: false,
			Message: "Insufficient permissions to change task status",
		}, domain.ErrInsufficientPermissions
	}

	// Validate status
	if !req.Status.IsValid() {
		return &domain.SuccessResponse{
			Success: false,
			Message: "Invalid task status",
		}, domain.ErrInvalidTaskStatus
	}

	// Change status
	_, err = s.taskRepo.ChangeStatus(ctx, taskID, req)
	if err != nil {
		return &domain.SuccessResponse{
			Success: false,
			Message: "Failed to change task status",
		}, err
	}

	return &domain.SuccessResponse{
		Success: true,
		Message: "Task status changed successfully",
	}, nil
}

// GetTasksByProject implements ports.TaskService
func (s *TaskService) GetTasksByProject(ctx context.Context, projectID uuid.UUID, filter *domain.TaskFilterRequest, userID uuid.UUID) (*domain.TaskListResponse, error) {
	// TODO: Add project permission check here - verify user has access to project

	// Apply default pagination if not provided
	s.applyDefaultPagination(filter)

	// Get tasks
	tasks, pagination, err := s.taskRepo.GetTasksByProject(ctx, projectID, filter)
	if err != nil {
		return &domain.TaskListResponse{
			Success: false,
			Data: domain.TaskListData{
				Tasks:      []domain.Task{},
				Pagination: domain.PaginationInfo{},
			},
		}, err
	}

	// Get summary statistics
	summary, err := s.taskRepo.GetTaskSummary(ctx, &projectID, nil)
	if err != nil {
		// Don't fail the request if summary fails, just log and continue
		summary = nil
	}

	return &domain.TaskListResponse{
		Success: true,
		Data: domain.TaskListData{
			Tasks:      tasks,
			Pagination: *pagination,
			Summary:    summary,
		},
	}, nil
}

// GetTasksByUser implements ports.TaskService
func (s *TaskService) GetTasksByUser(ctx context.Context, targetUserID uuid.UUID, filter *domain.TaskFilterRequest, requestUserID uuid.UUID) (*domain.TaskListResponse, error) {
	// Basic permission check - users can view their own tasks
	// TODO: Add more sophisticated project/team-based permission checks
	if targetUserID != requestUserID {
		// For now, allow viewing other users' tasks (could be restricted based on project membership)
	}

	// Apply default pagination if not provided
	s.applyDefaultPagination(filter)

	// Get tasks
	tasks, pagination, err := s.taskRepo.GetTasksByUser(ctx, targetUserID, filter)
	if err != nil {
		return &domain.TaskListResponse{
			Success: false,
			Data: domain.TaskListData{
				Tasks:      []domain.Task{},
				Pagination: domain.PaginationInfo{},
			},
		}, err
	}

	// Get summary statistics
	summary, err := s.taskRepo.GetTaskSummary(ctx, nil, &targetUserID)
	if err != nil {
		// Don't fail the request if summary fails
		summary = nil
	}

	return &domain.TaskListResponse{
		Success: true,
		Data: domain.TaskListData{
			Tasks:      tasks,
			Pagination: *pagination,
			Summary:    summary,
		},
	}, nil
}

// BulkUpdateTasks implements ports.TaskService
func (s *TaskService) BulkUpdateTasks(ctx context.Context, req *domain.BulkUpdateTasksRequest, userID uuid.UUID) (*domain.TaskListResponse, error) {
	// Validate request
	if err := s.validateBulkUpdateRequest(req); err != nil {
		return &domain.TaskListResponse{
			Success: false,
			Data: domain.TaskListData{
				Tasks:      []domain.Task{},
				Pagination: domain.PaginationInfo{},
			},
		}, err
	}

	// Check permissions for each task
	for index, taskID := range req.TaskIDs {
		task, err := s.taskRepo.GetByID(ctx, taskID)
		if err != nil {
			return nil, err
		}
		permCtx := &domain.PermissionContext{
			UserID:          userID,
			IsAuthenticated: true,
			TaskCreatorID:   &task.CreatedByUserID,
			TaskAssigneeID:  task.AssignedToUserID,
		}

		if !permCtx.CanEditTask() {
			fmt.Printf("removed")
			req.TaskIDs = append(req.TaskIDs[:index], req.TaskIDs[index+1:]...) // remove id
		}
	}

	// Perform bulk update
	updatedTasks, err := s.taskRepo.BulkUpdate(ctx, req)
	if err != nil {
		return &domain.TaskListResponse{
			Success: false,
			Data: domain.TaskListData{
				Tasks:      []domain.Task{},
				Pagination: domain.PaginationInfo{},
			},
		}, err
	}

	return &domain.TaskListResponse{
		Success: true,
		Data: domain.TaskListData{
			Tasks: updatedTasks,
			Pagination: domain.PaginationInfo{
				Page:       1,
				PageSize:   len(updatedTasks),
				TotalCount: len(updatedTasks),
				TotalPages: 1,
			},
		},
	}, nil
}

// GetTaskStatistics implements ports.TaskService
func (s *TaskService) GetTaskStatistics(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (*domain.TaskStatistics, error) {
	// Check if task exists and user has permission to view it
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Check permissions
	permCtx := &domain.PermissionContext{
		UserID:          userID,
		IsAuthenticated: true,
		TaskCreatorID:   &task.CreatedByUserID,
		TaskAssigneeID:  task.AssignedToUserID,
	}

	if !s.canViewTask(permCtx) {
		return nil, domain.ErrUnauthorized
	}

	// Get task statistics
	stats, err := s.taskRepo.GetTaskStatistics(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// Helper methods

func (s *TaskService) validateCreateTaskRequest(req *domain.CreateTaskRequest) error {
	if req.Title == "" {
		return domain.ErrRequiredFieldMissing
	}

	if req.EstimatedHours != nil && *req.EstimatedHours <= 0 {
		return domain.ErrEstimatedHoursInvalid
	}

	if req.ActualHours != nil && (*req.ActualHours < 0 || *req.ActualHours > 999) {
		return domain.ErrActualHoursInvalid
	}

	if req.Status != nil && !req.Status.IsValid() {
		return domain.ErrInvalidTaskStatus
	}

	if req.Priority != nil && !req.Priority.IsValid() {
		return domain.ErrInvalidTaskPriority
	}

	if req.DueDate != nil && req.DueDate.Before(time.Now()) {
		return domain.ErrDueDateInPast
	}

	return nil
}

func (s *TaskService) validateUpdateTaskRequest(req *domain.UpdateTaskRequest) error {
	if req.Title != nil && *req.Title == "" {
		return domain.ErrRequiredFieldMissing
	}

	if req.EstimatedHours != nil && *req.EstimatedHours <= 0 {
		return domain.ErrEstimatedHoursInvalid
	}

	if req.ActualHours != nil && (*req.ActualHours < 0 || *req.ActualHours > 999) {
		return domain.ErrActualHoursInvalid
	}

	if req.Status != nil && !req.Status.IsValid() {
		return domain.ErrInvalidTaskStatus
	}

	if req.Priority != nil && !req.Priority.IsValid() {
		return domain.ErrInvalidTaskPriority
	}

	return nil
}

func (s *TaskService) validateBulkUpdateRequest(req *domain.BulkUpdateTasksRequest) error {
	if len(req.TaskIDs) == 0 {
		return domain.ErrNoTasksSelected
	}

	if len(req.TaskIDs) > 100 {
		return domain.ErrTooManyTasksSelected
	}

	// Check if at least one update field is provided
	if req.Status == nil && req.Priority == nil && req.AssignedToUserID == nil && req.DueDate == nil {
		return domain.ErrNoUpdateFieldsProvided
	}

	if req.Status != nil && !req.Status.IsValid() {
		return domain.ErrInvalidTaskStatus
	}

	if req.Priority != nil && !req.Priority.IsValid() {
		return domain.ErrInvalidTaskPriority
	}

	return nil
}

func (s *TaskService) applyDefaultPagination(filter *domain.TaskFilterRequest) {
	if filter.Page == nil {
		page := 1
		filter.Page = &page
	}
	if filter.PageSize == nil {
		pageSize := 50
		filter.PageSize = &pageSize
	}
	
	// Validate pagination
	if *filter.Page < 1 {
		*filter.Page = 1
	}
	if *filter.PageSize < 1 || *filter.PageSize > 100 {
		*filter.PageSize = 50
	}
}

func (s *TaskService) canViewTask(permCtx *domain.PermissionContext) bool {
	if !permCtx.IsAuthenticated {
		return false
	}
	
	// Allow creator, assignee, or any authenticated user (for now)
	// TODO: Implement proper project/team-based permissions
	return true
}