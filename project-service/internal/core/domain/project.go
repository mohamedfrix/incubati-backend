package domain

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	Title              string    `json:"title" db:"title"`
	Description        *string   `json:"description" db:"description"`
	Domain             string    `json:"domain" db:"domain"`
	Status             string    `json:"status" db:"status"` // planning, active, on_hold, completed, cancelled
	StartDate          time.Time `json:"start_date" db:"start_date"`
	EndDate            time.Time `json:"end_date" db:"end_date"`
	ProgressPercentage int       `json:"progress_percentage" db:"progress_percentage"`
	IsPublic           bool      `json:"isPublic" db:"isPublic"`
	OwnerID            uuid.UUID `json:"owner_id" db:"owner_id"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy          uuid.UUID `json:"created_by" db:"created_by"`
	UpdatedBy          *uuid.UUID `json:"updated_by" db:"updated_by"`
}
