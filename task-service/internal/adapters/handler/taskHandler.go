package handler

import (
	"moulaybdl/zindy/task_service/internal/core/domain"
	"moulaybdl/zindy/task_service/internal/core/ports"
	"moulaybdl/zindy/task_service/pkg/utils"
	"net/http"
)



type TaskHandler struct {
	taskService ports.TaskService
}


func NewTaskHandler(taskService ports.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}	


func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var task_input domain.CreateTaskRequest

	err := utils.ReadJSON(w, r, &task_input)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "Invalid request body"}, nil)
		return
	}

	var task domain.Task

	task.Title = task_input.Title
	task.Description = task_input.Description

	var taskStatus domain.TaskStatus
	taskStatus = *task_input.Status
	if !taskStatus.IsValid() {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "Invalid task status"}, nil)
		return
	}

	var taskPriority domain.TaskPriority
	taskPriority = *task_input.Priority
	if !taskPriority.IsValid() {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "Invalid task priority"}, nil)
		return
	}

	due_date := task_input.DueDate
	if err != nil{
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "Invalid due date format"}, nil)
		return
	}
	task.DueDate = due_date

	comp_date := task_input.CompletedAt
	if err != nil{
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "Invalid complete date format"}, nil)
		return
	}
	task.CompletedAt = comp_date

	task.AssigneeID = *task_input.AssigneeID
	task.ProjectID = task_input.ProjectID
	task.MilestoneID = task_input.MilestoneID

	// set defaults:
	task.SetDefaults()

	// insert in the database
	h.taskService.CreateTask(r.Context(), &task_input)

}