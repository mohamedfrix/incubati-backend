package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProjectMentor struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	ProjectID       uuid.UUID  `json:"project_id" db:"project_id"`
	MentorID        uuid.UUID  `json:"mentor_id" db:"mentor_id"`
	MentorshipType  string     `json:"mentorship_type" db:"mentorship_type"` // technical, business, general, specialized
	Status          string     `json:"status" db:"status"`                   // pending, active, completed, cancelled
	StartDate       time.Time  `json:"start_date" db:"start_date"`
	EndDate         *time.Time `json:"end_date" db:"end_date"`
	HoursCommitted  int        `json:"hours_committed" db:"hours_committed"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	AssignedBy      uuid.UUID  `json:"assigned_by" db:"assigned_by"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type AddProjectMentorRequest struct {
	UserID         uuid.UUID `json:"assignee_id"`
	MentorID       uuid.UUID `json:"mentor_id"`
	MentorshipType string    `json:"mentorship_type"`
	StartDate      string    `json:"start_date"`
	EndDate        string    `json:"end_date"`
	HoursCommitted int       `json:"hours_committed"`
}
