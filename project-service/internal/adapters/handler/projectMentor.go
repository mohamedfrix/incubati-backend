package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
	"github.com/moulaybdl/incubAT/project_service/pkg/utils"
)


type ProjectMentorHandler struct {
	ProjectMentorService ports.ProjectMentorService
	ProjectMemberService ports.ProjectMemberService
	ProjectService ports.ProjectService
}



func NewProjectMentorHandler(projectMentorService ports.ProjectMentorService, ProjectMemberService  ports.ProjectMemberService, ProjectService ports.ProjectService) *ProjectMentorHandler {
	return &ProjectMentorHandler{
		ProjectMentorService: projectMentorService,
		ProjectMemberService: ProjectMemberService,
		ProjectService: ProjectService,
	}
}	


func (h *ProjectMentorHandler) AddMentorToProject(w http.ResponseWriter, r *http.Request) {
	var input domain.AddProjectMentorRequest

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "invalid request body"}, nil)
		return
	}


	// read the project ID:
	projectID_str := utils.GetURLparams(r, "project_id")
	projectID_uuid, err := uuid.Parse(projectID_str)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "invalid project ID"}, nil)
		return
	}

	// check permission here:
	//! this is already checked in the service
	permissions, err := h.ProjectMemberService.CheckMemberPermissions(projectID_uuid, input.UserID)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": "failed to check permissions"}, nil)
		return
	}
	if !permissions.CanManageTasks {
		project, err := h.ProjectService.GetProjectByID(projectID_str, input.UserID)
		if err != nil {
			utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": "failed to get project"}, nil)
			return
		}
		if project.OwnerID != input.UserID {
			utils.WriteJSON(w, r, http.StatusUnauthorized, utils.Envelope{"error": "you are not allowed to assign a mentor to this project"}, nil)
			return
		}
	}


	// assign the mentor:
	pMentor, err := h.ProjectMentorService.AssignMentor(r.Context(), input, projectID_uuid)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": err.Error()}, nil)
		return
	}

	// return the response:
	utils.WriteJSON(w, r, http.StatusCreated, utils.Envelope{"success": true, "message": "Mentor assigned successfully", "project_mentor": pMentor}, nil)
}
