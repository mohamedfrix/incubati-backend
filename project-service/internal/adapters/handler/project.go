package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
	"github.com/moulaybdl/incubAT/project_service/internal/core/services"
	"github.com/moulaybdl/incubAT/project_service/pkg/utils"
)



type ProjectHandler struct {
	// define services here
	project_service *services.ProjectService
	projectMember_service *services.ProjectMemberService
}

func NewProjectHandler(pr_service *services.ProjectService, projectMember_service  *services.ProjectMemberService) *ProjectHandler{
	return &ProjectHandler{
		project_service: pr_service,
		projectMember_service: projectMember_service,
	}
}


func (p *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {

	var input domain.CreateProjectRequest
	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "Malformed JSON: could not parse request body"}, nil)
		return
	}

	// insert in the database:
	response , err := p.project_service.CreateProject(&input, input.UserID)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": err.Error()}, nil)
		return
	}

	// Auto-add creator as project manager with full permissions (as per documentation)
	permissions := ports.ProjectMemberPermissions{
		CanEditProject: true,
		CanManageTasks: true,
		CanViewReports: true,
	}
	
	_, err = p.projectMember_service.AddMemberToProject(response.ID, input.UserID, "manager", permissions)
	if err != nil {
		// Log the error but don't fail the project creation
		fmt.Printf("Warning: Failed to add creator as project manager: %v\n", err)
	}

	// return reponse:
	utils.WriteJSON(w, r, http.StatusCreated, utils.Envelope{"success": true, "message":"Project created successfully" ,"data": response}, nil)

}

func (p *ProjectHandler) GetProjectByID(w http.ResponseWriter, r *http.Request) {
	projectID_str := utils.GetURLparams(r, "project_id")

	//! input: user should include his id to check the permission:
	//! here i will assume that the id is incldued as url params
	userID_str := utils.GetURLparams(r, "user_id")
	userID_uuid, err := uuid.Parse(userID_str)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "Invalid user ID format"}, nil)
		return
	}

	project, err := p.project_service.GetProjectByID(projectID_str, userID_uuid)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": err.Error()}, nil)
		return
	}
	utils.WriteJSON(w, r, http.StatusOK, utils.Envelope{"success": true, "message":"Project retrieved successfully" ,"project": project}, nil)
}

func (p *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	var input domain.UpdateProjectRequest

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "Malformed JSON: could not parse request body"}, nil)
		return
	}

	// retrieve the project id:
	projectID_str := utils.GetURLparams(r, "project_id")
	projectID_uuid, err := uuid.Parse(projectID_str)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "Invalid project ID format"}, nil)
		return
	}

	// update the project:
	new_project, err := p.project_service.UpdateProject(&input, input.UserID, projectID_uuid)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": err.Error()}, nil)
		return
	}

	// return reponse:
		utils.WriteJSON(w, r, http.StatusOK, utils.Envelope{"success": true, "message":"Project updated successfully" ,"project": new_project}, nil)
}


func (p *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	projectID_str := utils.GetURLparams(r, "project_id")

	var input struct {
		UserID uuid.UUID `json:"user_id"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "Malformed JSON: could not parse request body"}, nil)
		return
	}

	title, err := p.project_service.DeleteProject(projectID_str, input.UserID)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": err.Error()}, nil)
		return
	}

	// return sucess status:
	utils.WriteJSON(w, r, http.StatusOK, utils.Envelope{"success": true, "message": fmt.Sprintf("Project %s deleted successfully", *title), }, nil)
}


func (p *ProjectHandler) GetProjectStatistics(w http.ResponseWriter, r *http.Request) {
	projectID_str := utils.GetURLparams(r, "project_id")
	
	projectID, err := uuid.Parse(projectID_str)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "Invalid project ID format"}, nil)
		return
	}

	// check permission:
	// get the user id:
	userID_str := r.Header.Get("user_id")
	userID_uuid, err := uuid.Parse(userID_str)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "Invalid user ID format"}, nil)
		return
	}

	// check if the project is public
	project, err := p.project_service.GetProjectByID(projectID_str, userID_uuid )
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": err.Error()}, nil)
		return
	}

	if !project.IsPublic {
		// check if the user is a member of the project
		permission, err := p.projectMember_service.CheckMemberPermissions(project.ID, userID_uuid)
		if err != nil {
			utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": err.Error()}, nil)
			return
		}

		if !permission.IsMember {
			utils.WriteJSON(w, r, http.StatusUnauthorized, utils.Envelope{"error": "You do not have permission to access this project's statistics"}, nil)
			return
		}
	}

	// Get project statistics
	statistics, err := p.project_service.GetProjectStatistics(r.Context(), projectID)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": err.Error()}, nil)
		return
	}

	// Return success response
	utils.WriteJSON(w, r, http.StatusOK, utils.Envelope{
		"success": true,
		"data":    statistics,
	}, nil)
}


func (p *ProjectHandler) GetAllProjects(w http.ResponseWriter, r *http.Request) {
	//! once again, lets assume the user_id is included as URL params:
	userID_str := utils.GetURLparams(r, "user_id") // This is just to show that we assume the user_id is included
	userID_uuid, err := uuid.Parse(userID_str)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "Invalid user ID format"}, nil)
		return
	}

	// Parse query parameters for filtering and pagination
	query := r.URL.Query()
	
	// Default pagination values
	page := 1
	pageSize := 10
	
	// Parse page
	if pageStr := query.Get("page"); pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}
	
	// Parse page_size
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		if parsedPageSize, err := strconv.Atoi(pageSizeStr); err == nil && parsedPageSize > 0 && parsedPageSize <= 100 {
			pageSize = parsedPageSize
		}
	}
	
	// Calculate offset
	offset := (page - 1) * pageSize
	
	// Build filters map
	filters := make(map[string]interface{})
	
	if status := query.Get("status"); status != "" {
		filters["status"] = status
	}
	
	if domain := query.Get("domain"); domain != "" {
		filters["domain"] = domain
	}
	
	if search := query.Get("search"); search != "" {
		filters["search"] = search
	}
	
	if orderBy := query.Get("order_by"); orderBy != "" {
		filters["order_by"] = orderBy
	} else {
		filters["order_by"] = "-created_at" // Default order
	}
	
	// Get projects from service
	response, err := p.project_service.GetAllProjects(pageSize, offset, filters, userID_uuid)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": err.Error()}, nil)
		return
	}
		
	// Update the response with filtered projects and adjust pagination
	response.Pagination.TotalCount = len(response.Projects)
	response.Pagination.TotalPages = (len(response.Projects) + pageSize - 1) / pageSize // Ceiling division
	
	// Return success response
	utils.WriteJSON(w, r, http.StatusOK, utils.Envelope{
		"success": true,
		"data":    response,
	}, nil)
}