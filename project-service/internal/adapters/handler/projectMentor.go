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
}



func NewProjectMentorHandler(projectMentorService ports.ProjectMentorService) *ProjectMentorHandler {
	return &ProjectMentorHandler{
		ProjectMentorService: projectMentorService,
	}
}	


func (h *ProjectMentorHandler) AddMentorToProject(w http.ResponseWriter, r *http.Request) {
	var input struct {
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

	// create projectMentor object:
	var projectMentor domain.ProjectMentor
	mentorID, err := uuid.Parse(input.Mentor)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "invalid mentor ID"}, nil)
		return
	}
	projectMentor.ID = mentorID
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

	projectMentor.ProjectID, err = uuid.Parse(projectID_str)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "invalid project ID"}, nil)
		return
	}

	// assign the mentor:
	pMentor, err := h.ProjectMentorService.AssignMentor(r.Context(), &projectMentor)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": err.Error()}, nil)
		return
	}

	// return the response:
	utils.WriteJSON(w, r, http.StatusCreated, utils.Envelope{"success": true, "message": "Mentor assigned successfully", "project_mentor": pMentor}, nil)
}
