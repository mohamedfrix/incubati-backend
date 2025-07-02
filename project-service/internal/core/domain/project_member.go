package domain

import (
	"time"

	"github.com/google/uuid"
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
