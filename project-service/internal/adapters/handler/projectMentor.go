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
		UserID uuid.UUID `json:"assignee_id"`
		Mentor string `json:"mentor"`
		Mentorship string `json:"mentorship"`
		StartDate string `json:"start_date"`
		EndDate string `json:"end_date"`
		HoursCommitted int `json:"hours_committed"`
	}

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


	// create projectMentor object:
	var projectMentor domain.ProjectMentor
	mentorID, err := uuid.Parse(input.Mentor)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "invalid mentor ID"}, nil)
		return
	}
	projectMentor.MentorID = mentorID
	projectMentor.MentorshipType = input.Mentorship

	s_date, err := utils.FromStringToTime(input.StartDate)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "invalid start date format"}, nil)
		return
	}
	projectMentor.StartDate = *s_date

	e_date, err := utils.FromStringToTime(input.EndDate)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "invalid end date format"}, nil)
		return
	}
	projectMentor.EndDate = e_date

	projectMentor.HoursCommitted = input.HoursCommitted

	projectMentor.ProjectID = projectID_uuid

	projectMentor.AssignedBy = input.UserID


	// assign the mentor:
	pMentor, err := h.ProjectMentorService.AssignMentor(r.Context(), &projectMentor)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": err.Error()}, nil)
		return
	}

	// return the response:
	utils.WriteJSON(w, r, http.StatusCreated, utils.Envelope{"success": true, "message": "Mentor assigned successfully", "project_mentor": pMentor}, nil)
}
