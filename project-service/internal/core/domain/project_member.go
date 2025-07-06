package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/pkg/utils"
)

type ProjectMember struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	ProjectID      uuid.UUID  `json:"project_id" db:"project_id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	Role           string     `json:"role" db:"role"` // manager, lead, developer, designer, analyst, tester, stakeholder
	IsActive       bool       `json:"is_active" db:"is_active"`
	JoinedDate     time.Time  `json:"joined_date" db:"joined_date"`
	LeftDate       *time.Time `json:"left_date" db:"left_date"`
	CanEditProject bool       `json:"can_edit_project" db:"can_edit_project"`
	CanManageTasks bool       `json:"can_manage_tasks" db:"can_manage_tasks"`
	CanViewReports bool       `json:"can_view_reports" db:"can_view_reports"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}


type ProjectMemberValidator struct {
	 	Errors []map[string]string
}

func (p *ProjectMemberValidator) CheckValid() bool {
	return len(p.Errors) == 0
}


func (p *ProjectMemberValidator) Validate(pm *ProjectMember) {
	// check if value of role is valid
	valid  := []string{"manager", "lead", "developer", "designer", "analyst", "tester", "stakeholder"}
	if ok := utils.CheckContains(valid, pm.Role); !ok {
		p.Errors = append(p.Errors, map[string]string{"role": "not a valid value"})
		return
	}
}

type AddProjectMemberRequest struct {
		AssigneeID uuid.UUID `json:"assignee_id"` // the user who is adding the member
		UserID uuid.UUID `json:"user_id"`
		Role string `json:"role"`
		Can_edit_project bool `json:"can_edit_project"`
		Can_manage_tasks bool `json:"can_manage_tasks"`
		Can_view_reports bool `json:"can_view_reports"`
}