package handler

import (
	"moulaybdl/zindy/task_service/internal/core/domain"
	"moulaybdl/zindy/task_service/internal/core/ports"
	"moulaybdl/zindy/task_service/pkg/utils"
	"net/http"

	"github.com/google/uuid"
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

	// insert in the database
	created_task, err := h.taskService.CreateTask(r.Context(), &task_input)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": err.Error()}, nil)
		return
	}


	// return reponse:
	utils.WriteJSON(w, r, http.StatusCreated, utils.Envelope{"success": true, "message":"task created", "task": created_task}, nil)
}


func (h *TaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	task_id := utils.GetURLparams(r, "task_id")
	taskID_uuid, err := uuid.Parse(task_id)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "malformat for uuid"}, nil)
		return
	}


	task, err := h.taskService.GetTask(r.Context(), taskID_uuid)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": err.Error()}, nil)
		return
	}

	// return response:
	utils.WriteJSON(w, r, http.StatusOK, utils.Envelope{"success": true, "task": task}, nil)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	var input domain.UpdateTaskRequest
	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "Invalid request body"}, nil)
		return
	}

	task_id := utils.GetURLparams(r, "task_id")
	taskID_uuid, err := uuid.Parse(task_id)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "malformat for uuid"}, nil)
		return
	}

	// update the task in the database:
	update_task, err := h.taskService.UpdateTask(r.Context(), taskID_uuid, &input)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": err.Error()}, nil)
		return
	}

	// return response:
	utils.WriteJSON(w, r, http.StatusOK, utils.Envelope{"success": true, "message": "task updated", "task": update_task}, nil)
}


func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskID := utils.GetURLparams(r, "task_id")
	taskID_uuid, err := uuid.Parse(taskID)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "malformat for uuid"}, nil)
		return
	}

	err = h.taskService.DeleteTask(r.Context(), taskID_uuid)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": err.Error()}, nil)
		return
	}

	// return reponse:
	utils.WriteJSON(w, r, http.StatusOK, utils.Envelope{"success": true, "message": "task deleted"}, nil)

}