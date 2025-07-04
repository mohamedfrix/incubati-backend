package services

import (
	"context"
	"moulaybdl/zindy/task_service/internal/core/domain"
	"moulaybdl/zindy/task_service/internal/core/ports"
)



type TaskService struct {
	TaskRepo ports.TaskRepository
}


func NewTaskService(task_repo ports.TaskRepository) *TaskService {
	return &TaskService{
		TaskRepo: task_repo,
	}
}


func (t *TaskService) CreateTask(req *domain.CreateTaskRequest) (*domain.Task, error) {
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

