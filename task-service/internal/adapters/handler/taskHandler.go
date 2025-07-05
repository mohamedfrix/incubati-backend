package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"moulaybdl/zindy/task_service/internal/core/domain"
	"moulaybdl/zindy/task_service/internal/core/ports"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

type TaskHandler struct {
	taskService    ports.TaskService
	commentService ports.CommentService
}

func NewTaskHandler(taskService ports.TaskService, commentService ports.CommentService) *TaskHandler {
	return &TaskHandler{
		taskService:    taskService,
		commentService: commentService,
	}
}

// CreateTask handles POST /api/tasks
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req domain.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload", err)
		return
	}

	//! assuming userID of the current user is store in the context
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required", err)
		return
	}

	response, err := h.taskService.CreateTask(r.Context(), &req, userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Failed to create task", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, response)
}

// GetTaskByID handles GET /api/tasks/{task_id}
func (h *TaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	taskID, err := uuid.Parse(ps.ByName("task_id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}

	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required", err)
		return
	}

	response, err := h.taskService.GetTaskByID(r.Context(), taskID, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrTaskNotFound {
			status = http.StatusNotFound
		} else if err == domain.ErrUnauthorized {
			status = http.StatusForbidden
		}
		h.writeErrorResponse(w, status, "Failed to get task", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// UpdateTask handles PUT/PATCH /api/tasks/{task_id}
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	taskID, err := uuid.Parse(ps.ByName("task_id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}

	var req domain.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload", err)
		return
	}

	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required", err)
		return
	}

	// userID, _ = uuid.Parse("123e4567-e89b-12d3-a456-426614174003")

	response, err := h.taskService.UpdateTask(r.Context(), taskID, &req, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrTaskNotFound {
			status = http.StatusNotFound
		} else if err == domain.ErrInsufficientPermissions {
			status = http.StatusForbidden
		}
		h.writeErrorResponse(w, status, "Failed to update task", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// DeleteTask handles DELETE /api/tasks/{task_id}
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	taskID, err := uuid.Parse(ps.ByName("task_id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}

	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required", err)
		return
	}


	response, err := h.taskService.DeleteTask(r.Context(), taskID, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrTaskNotFound {
			status = http.StatusNotFound
		} else if err == domain.ErrNotTaskCreator || err == domain.ErrInsufficientPermissions {
			status = http.StatusForbidden
		} else if err == domain.ErrTaskHasSubtasks {
			status = http.StatusConflict
		}
		h.writeErrorResponse(w, status, "Failed to delete task", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// AssignTask handles POST /api/tasks/{task_id}/assign
func (h *TaskHandler) AssignTask(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	taskID, err := uuid.Parse(ps.ByName("task_id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}

	var req domain.AssignTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload", err)
		return
	}

	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required", err)
		return
	}

	response, err := h.taskService.AssignTask(r.Context(), taskID, &req, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrTaskNotFound {
			status = http.StatusNotFound
		} else if err == domain.ErrInsufficientPermissions {
			status = http.StatusForbidden
		}
		h.writeErrorResponse(w, status, "Failed to assign task", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// UnassignTask handles DELETE /api/tasks/{task_id}/assign
func (h *TaskHandler) UnassignTask(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	taskID, err := uuid.Parse(ps.ByName("task_id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}

	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required", err)
		return
	}

	response, err := h.taskService.UnassignTask(r.Context(), taskID, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrTaskNotFound {
			status = http.StatusNotFound
		} else if err == domain.ErrInsufficientPermissions {
			status = http.StatusForbidden
		}
		h.writeErrorResponse(w, status, "Failed to unassign task", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// ChangeTaskStatus handles POST /api/tasks/{task_id}/status
func (h *TaskHandler) ChangeTaskStatus(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	taskID, err := uuid.Parse(ps.ByName("task_id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}

	var req domain.ChangeStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload", err)
		return
	}

	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required", err)
		return
	}


	response, err := h.taskService.ChangeTaskStatus(r.Context(), taskID, &req, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrTaskNotFound {
			status = http.StatusNotFound
		} else if err == domain.ErrInsufficientPermissions {
			status = http.StatusForbidden
		} else if err == domain.ErrInvalidTaskStatus {
			status = http.StatusBadRequest
		}
		h.writeErrorResponse(w, status, "Failed to change task status", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetTasksByProject handles GET /api/projects/{project_id}/tasks
func (h *TaskHandler) GetTasksByProject(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	projectID, err := uuid.Parse(ps.ByName("project_id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid project ID", err)
		return
	}

	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required", err)
		return
	}

	filter := h.parseTaskFilter(r)
	response, err := h.taskService.GetTasksByProject(r.Context(), projectID, filter, userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get tasks", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetTasksByUser handles GET /api/users/{user_id}/tasks
func (h *TaskHandler) GetTasksByUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	targetUserID, err := uuid.Parse(ps.ByName("user_id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required", err)
		return
	}

	filter := h.parseTaskFilter(r)
	response, err := h.taskService.GetTasksByUser(r.Context(), targetUserID, filter, userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get tasks", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// BulkUpdateTasks handles POST /api/tasks/bulk-update
func (h *TaskHandler) BulkUpdateTasks(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req domain.BulkUpdateTasksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload", err)
		return
	}

	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required", err)
		return
	}

	response, err := h.taskService.BulkUpdateTasks(r.Context(), &req, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrNoTasksSelected || err == domain.ErrTooManyTasksSelected || err == domain.ErrNoUpdateFieldsProvided {
			status = http.StatusBadRequest
		}
		h.writeErrorResponse(w, status, "Failed to bulk update tasks", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetTaskStatistics handles GET /api/tasks/{task_id}/statistics
func (h *TaskHandler) GetTaskStatistics(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	taskID, err := uuid.Parse(ps.ByName("task_id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}

	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required", err)
		return
	}

	statistics, err := h.taskService.GetTaskStatistics(r.Context(), taskID, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrTaskNotFound {
			status = http.StatusNotFound
		} else if err == domain.ErrUnauthorized {
			status = http.StatusForbidden
		}
		h.writeErrorResponse(w, status, "Failed to get task statistics", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, statistics)
}

// AddComment handles POST /api/tasks/{task_id}/comments
func (h *TaskHandler) AddComment(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	taskID, err := uuid.Parse(ps.ByName("task_id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}

	var req domain.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON payload", err)
		return
	}

	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required", err)
		return
	}

	response, err := h.commentService.AddComment(r.Context(), taskID, &req, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrTaskNotFound {
			status = http.StatusNotFound
		} else if err == domain.ErrUnauthorized {
			status = http.StatusForbidden
		} else if err == domain.ErrCommentContentEmpty || err == domain.ErrCommentTooLong {
			status = http.StatusBadRequest
		}
		h.writeErrorResponse(w, status, "Failed to add comment", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, response)
}

// GetTaskComments handles GET /api/tasks/{task_id}/comments
func (h *TaskHandler) GetTaskComments(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	taskID, err := uuid.Parse(ps.ByName("task_id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}

	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required", err)
		return
	}

	pagination := h.parsePagination(r)
	response, err := h.commentService.GetTaskComments(r.Context(), taskID, pagination, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == domain.ErrTaskNotFound {
			status = http.StatusNotFound
		} else if err == domain.ErrUnauthorized {
			status = http.StatusForbidden
		}
		h.writeErrorResponse(w, status, "Failed to get comments", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// Helper methods

func (h *TaskHandler) getUserIDFromContext(r *http.Request) (uuid.UUID, error) {
	// TODO: Implement proper authentication middleware that sets user ID in context
	// For now, return a dummy user ID
	return uuid.New(), nil
}

func (h *TaskHandler) parseTaskFilter(r *http.Request) *domain.TaskFilterRequest {
	filter := &domain.TaskFilterRequest{}
	
	if status := r.URL.Query().Get("status"); status != "" {
		taskStatus := domain.TaskStatus(status)
		filter.Status = &taskStatus
	}
	
	if priority := r.URL.Query().Get("priority"); priority != "" {
		taskPriority := domain.TaskPriority(priority)
		filter.Priority = &taskPriority
	}
	
	if assignedTo := r.URL.Query().Get("assigned_to_user_id"); assignedTo != "" {
		if userID, err := uuid.Parse(assignedTo); err == nil {
			filter.AssignedToUserID = &userID
		}
	}
	
	if projectID := r.URL.Query().Get("project_id"); projectID != "" {
		if id, err := uuid.Parse(projectID); err == nil {
			filter.ProjectID = &id
		}
	}
	
	if search := r.URL.Query().Get("search"); search != "" {
		filter.Search = &search
	}
	
	if includeSubtasks := r.URL.Query().Get("include_subtasks"); includeSubtasks == "true" {
		val := true
		filter.IncludeSubtasks = &val
	}
	
	if overdueOnly := r.URL.Query().Get("overdue_only"); overdueOnly == "true" {
		val := true
		filter.OverdueOnly = &val
	}
	
	if dueSoon := r.URL.Query().Get("due_soon"); dueSoon != "" {
		if days, err := strconv.Atoi(dueSoon); err == nil {
			filter.DueSoon = &days
		}
	}
	
	if orderBy := r.URL.Query().Get("order_by"); orderBy != "" {
		filter.OrderBy = &orderBy
	}
	
	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filter.Page = &p
		}
	}
	
	if pageSize := r.URL.Query().Get("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			filter.PageSize = &ps
		}
	}
	
	return filter
}

func (h *TaskHandler) parsePagination(r *http.Request) *domain.PaginationInfo {
	pagination := &domain.PaginationInfo{
		Page:     1,
		PageSize: 20,
	}
	
	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			pagination.Page = p
		}
	}
	
	if pageSize := r.URL.Query().Get("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
			pagination.PageSize = ps
		}
	}
	
	return pagination
}

func (h *TaskHandler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *TaskHandler) writeErrorResponse(w http.ResponseWriter, status int, message string, err error) {
	errorResp := domain.ErrorResponse{
		Success: false,
		Message: message,
		Error:   err.Error(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResp)
}