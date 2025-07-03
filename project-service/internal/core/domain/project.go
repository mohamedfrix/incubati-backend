package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/pkg/utils"
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

type ProjectValidator struct {
 	Errors []map[string]string
}

func (p *ProjectValidator) CheckValid() bool {
	return len(p.Errors) == 0
}



func (p *ProjectValidator) Validate(pr *Project) {
	// check status values
	valid := []string {"planning", "active", "on_hold", "completed", "cancelled"}
	if ok := utils.CheckContains(valid, pr.Status); !ok {
		p.Errors = append(p.Errors, map[string]string{"status": "not a valid value"})
		return 
	}

	// check percentage:
	if pr.ProgressPercentage < 0 || pr.ProgressPercentage > 100 {
		p.Errors = append(p.Errors, map[string]string{"percentage": "invalid value"})
		return
	}

}


