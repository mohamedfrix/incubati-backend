package domain

import (
	"time"

	"github.com/google/uuid"
)

type Milestone struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	ProjectID          uuid.UUID  `json:"project_id" db:"project_id"`
	Title              string     `json:"title" db:"title"`
	Description        *string    `json:"description" db:"description"`
	DueDate            time.Time  `json:"due_date" db:"due_date"`
	CompletedDate      *time.Time `json:"completed_date" db:"completed_date"`
	IsCompleted        bool       `json:"isCompleted" db:"isCompleted"`
	Status             string     `json:"status" db:"status"` // pending, in_progress, completed, overdue
	ProgressPercentage int        `json:"progress_percentage" db:"progress_percentage"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy          uuid.UUID  `json:"created_by" db:"created_by"`
}
