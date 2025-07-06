package handler

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
	"github.com/moulaybdl/incubAT/project_service/pkg/utils"
)



type ProjectMemberHandler struct {
	// services here
	ProjectMemberService ports.ProjectMemberService
	ProjectService ports.ProjectService
}

func NewProjectMemberHandler(projectMemberService ports.ProjectMemberService, ProjectService ports.ProjectService) *ProjectMemberHandler {
	return &ProjectMemberHandler{
		ProjectMemberService: projectMemberService,
		ProjectService: ProjectService,
	}
}


func (p *ProjectMemberHandler) AddMemberToProject(w http.ResponseWriter, r *http.Request) {	
	var input struct {
		// add data about the send of the request
		AssigneeID uuid.UUID `json:"assignee_id"` // the user who is adding the member
		UserID uuid.UUID `json:"user_id"`
		Role string `json:"role"`
		Can_edit_project bool `json:"can_edit_project"`
		Can_manage_tasks bool `json:"can_manage_tasks"`
		Can_view_reports bool `json:"can_view_reports"`
	}

	err := utils.ReadJSON(w, r, &input)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"status": fmt.Sprintf("Invalid json format: %s", err.Error())}, nil)
		return
	}

	// get the URL params
	projectID_str := utils.GetURLparams(r, "project_id")

	projectID_uuid, err := uuid.Parse(projectID_str)
	if err != nil {	
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"status": "Invalid project ID format"}, nil)
		return
	}

	// check if the user is allowed to add a member
	project, err := p.ProjectService.GetProjectByID(projectID_str, input.AssigneeID)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusNotFound, utils.Envelope{"status": fmt.Sprintf("Project not found: %s", err.Error())}, nil)
		return
	}

	if project == nil || project.OwnerID != input.AssigneeID {
		// check if the user has permission to add a member
		permission, err := p.ProjectMemberService.CheckMemberPermissions(projectID_uuid, input.AssigneeID)
		if err != nil {
			utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"status": fmt.Sprintf("Failed to check member permissions: %s", err.Error())}, nil)
			return
		}

		if !permission.CanManageTasks {
			utils.WriteJSON(w, r, http.StatusUnauthorized, utils.Envelope{"status": "You do not have permission to add members to this project"}, nil)
			return
		}
	}


	var member domain.ProjectMember
	member.Role = input.Role

	// validate:
	var validator domain.ProjectMemberValidator
	validator.Validate(&member)

	if !validator.CheckValid(){
		utils.WriteJSON(w, r, http.StatusBadRequest, utils.Envelope{"status": "Invalid project member data", "errors": validator.Errors}, nil)
		return
	}


	var permissions ports.ProjectMemberPermissions
	permissions.CanEditProject = input.Can_edit_project
	permissions.CanManageTasks = input.Can_manage_tasks
	permissions.CanViewReports = input.Can_view_reports

	new_member, err := p.ProjectMemberService.AddMemberToProject(projectID_uuid, input.UserID, input.Role, permissions)
	if err != nil {
		utils.WriteJSON(w, r, http.StatusInternalServerError, utils.Envelope{"status": fmt.Sprintf("Failed to add member to project: %s", err.Error())}, nil)
		return
	}

	utils.WriteJSON(w, r, http.StatusCreated, utils.Envelope{"success": true, "status": "Member added successfully", "member": new_member}, nil)

}


