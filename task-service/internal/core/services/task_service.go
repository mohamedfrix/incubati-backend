package services

import (
	"context"
	"moulaybdl/zindy/task_service/internal/core/domain"
	"moulaybdl/zindy/task_service/internal/core/ports"

	"github.com/google/uuid"
)



type TaskService struct {
	TaskRepo ports.TaskRepository
}


func NewTaskService(task_repo ports.TaskRepository) *TaskService {
	return &TaskService{
		TaskRepo: task_repo,
	}
}


func (t *TaskService) CreateTask(ctx context.Context, req *domain.CreateTaskRequest) (*domain.Task, error) {
		var task domain.Task

	task.Title = req.Title
	task.Description = req.Description

	var taskStatus domain.TaskStatus
	taskStatus = *req.Status
	if !taskStatus.IsValid() {
		return nil, domain.ErrInvalidTaskStatus
	}

	var taskPriority domain.TaskPriority
	taskPriority = *req.Priority
	if !taskPriority.IsValid() {
		return nil, domain.ErrInvalidPriority
	}

	due_date := req.DueDate
	task.DueDate = due_date

	comp_date := req.CompletedAt
	task.CompletedAt = comp_date

	task.AssigneeID = *req.AssigneeID
	task.ProjectID = req.ProjectID
	task.MilestoneID = req.MilestoneID

	// set defaults:
	task.SetDefaults()

	return t.TaskRepo.Create(context.Background(), &task)
}


func (t *TaskService) GetTask(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	return t.TaskRepo.GetByID(context.Background(), id)
}


func (t *TaskService) UpdateTask(ctx context.Context, id uuid.UUID, req *domain.UpdateTaskRequest) (*domain.Task, error) {
	return t.TaskRepo.Update(ctx, id, req)
}

func (t *TaskService) DeleteTask(ctx context.Context, id uuid.UUID) error {
	return t.TaskRepo.Delete(ctx, id)
}
