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
	var input struct {
		MentorID       string `json:"mentor_id"`
		MentorshipType string `json:"mentorship_type"`
		StartDate      string `json:"start_date"`
		EndDate        string `json:"end_date"`
		HoursCommitted int    `json:"hours_committed"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "invalid request body"}, nil)
		return
	}

	// Validate mentor ID
	mentorID, err := uuid.Parse(input.MentorID)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "Invalid mentor ID format"}, nil)
		return
	}

	// Validate required fields
	if input.MentorshipType == "" {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "mentorship_type is required"}, nil)
		return
	}

	if input.StartDate == "" {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "start_date is required"}, nil)
		return
	}

	// Get user ID from context (assuming auth middleware sets this)
	assigneeID, err := getUserIDFromContext(r.Context())
	if err != nil {
		utils.WriteJSON(w, r, http.StatusUnauthorized, utils.Envelope{"error": "User not authenticated"}, nil)
		return
	}

	// read the project ID:
	projectID_str := utils.GetURLparams(r, "project_id")
	projectID_uuid, err := uuid.Parse(projectID_str)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "invalid project ID"}, nil)
		return
	}

	// Convert to domain request
	domainInput := domain.AddProjectMentorRequest{
		UserID:         assigneeID,
		MentorID:       mentorID,
		MentorshipType: input.MentorshipType,
		StartDate:      input.StartDate,
		EndDate:        input.EndDate,
		HoursCommitted: input.HoursCommitted,
	}

	// check permission here:
	//! this is already checked in the service
	permissions, err := h.ProjectMemberService.CheckMemberPermissions(projectID_uuid, assigneeID)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": "failed to check permissions"}, nil)
		return
	}
	if !permissions.CanManageTasks {
		project, err := h.ProjectService.GetProjectByID(projectID_str, assigneeID)
		if err != nil {
			utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": "failed to get project"}, nil)
			return
		}
		if project.OwnerID != assigneeID {
			utils.WriteJSON(w, r, http.StatusUnauthorized, utils.Envelope{"error": "you are not allowed to assign a mentor to this project"}, nil)
			return
		}
	}


	// assign the mentor:
	pMentor, err := h.ProjectMentorService.AssignMentor(r.Context(), domainInput, projectID_uuid)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": err.Error()}, nil)
		return
	}

	// return the response:
	utils.WriteJSON(w, r, http.StatusCreated, utils.Envelope{"success": true, "message": "Mentor assigned successfully", "project_mentor": pMentor}, nil)
}
