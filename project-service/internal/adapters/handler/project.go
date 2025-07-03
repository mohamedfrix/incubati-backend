package handler

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
	"github.com/moulaybdl/incubAT/project_service/internal/core/services"
	"github.com/moulaybdl/incubAT/project_service/pkg/utils"
)



type ProjectHandler struct {
	// define services here
	project_service *services.ProjectService
}

func NewProjectHandler(pr_service *services.ProjectService) *ProjectHandler{
	return &ProjectHandler{
		project_service: pr_service,
	}
}


func (p *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Title string `json:"title"`
		Description string `json:"description"`
		Domain string `json:"domain"`
		Status string `json:"status"`
		StartDate string `json:"start_date"`
		EndDate string `json:"end_date"`
		ProgressPercentage int `json:"progress_percentage"`
		IsPublic bool `json:"is_public"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"error": "Malformed JSON: could not parse request body"}, nil)
		return
	}

	// initialize Project struct:
	var project domain.Project
	project.Title = input.Title
	project.Description = &input.Description
	project.Domain = input.Domain
	project.Status = input.Status

	s_date, err := utils.FromStringToTime(input.StartDate)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"date": "invalid date format"}, nil)
		return
	}
	project.StartDate = *s_date

	e_date, err := utils.FromStringToTime(input.StartDate)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"date": "invalid date format"}, nil)
		return
	}
	project.EndDate = *e_date

	project.ProgressPercentage = input.ProgressPercentage
	project.IsPublic = input.IsPublic


	// validate the input of needed 
	validator := domain.ProjectValidator{}
	validator.Validate(&project)

	// check if any errors:

	if !validator.CheckValid() {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"errors": validator.Errors}, nil)
		return
	}

	// insert in the database:
	response , err := p.project_service.CreateProject(&project)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": err.Error()}, nil)
		return
	}

	// return reponse:
	utils.WriteJSON(w, r, http.StatusCreated, utils.Envelope{"sucess": true, "message":"Project created successfully" ,"project": response}, nil)

}


func (p *ProjectHandler) GetProjectByID(w http.ResponseWriter, r *http.Request) {
	projectID_str := utils.GetURLparams(r, "project_id")

	project, err := p.project_service.GetProjectByID(projectID_str)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": err.Error()}, nil)
		return
	}

	utils.WriteJSON(w, r, http.StatusOK, utils.Envelope{"sucess": true, "message":"Project retrieved successfully" ,"project": project}, nil)
}

func (p *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
		Description string `json:"description"`
		Domain string `json:"domain"`
		Status string `json:"status"`
		StartDate string `json:"start_date"`
		EndDate string `json:"end_date"`
		ProgressPercentage int `json:"progress_percentage"`
		IsPublic bool `json:"is_public"`
	}

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

	// initialize Project struct:
	var project domain.Project
	project.ID = projectID_uuid
	project.Title = input.Title
	project.Description = &input.Description
	project.Domain = input.Domain
	project.Status = input.Status

	s_date, err := utils.FromStringToTime(input.StartDate)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"date": "invalid date format"}, nil)
		return
	}
	project.StartDate = *s_date

	e_date, err := utils.FromStringToTime(input.StartDate)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"date": "invalid date format"}, nil)
		return
	}
	project.EndDate = *e_date

	project.ProgressPercentage = input.ProgressPercentage
	project.IsPublic = input.IsPublic


	// validate the input of needed 
	validator := domain.ProjectValidator{}
	validator.Validate(&project)

	// check if any errors:

	if !validator.CheckValid() {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"errors": validator.Errors}, nil)
		return
	}

	// update the project:
	new_project, err := p.project_service.UpdateProject(&project)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": err.Error()}, nil)
		return
	}


	// return reponse:
		utils.WriteJSON(w, r, http.StatusCreated, utils.Envelope{"sucess": true, "message":"Project updated successfully" ,"project": new_project}, nil)

}


func (p *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	projectID_str := utils.GetURLparams(r, "project_id")

	title, err := p.project_service.DeleteProject(projectID_str)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"error": err.Error()}, nil)
		return
	}

	// return sucess status:
	utils.WriteJSON(w, r, http.StatusOK, utils.Envelope{"success": true, "message": fmt.Sprintf("Project %s deleted successfully", *title), }, nil)

} 