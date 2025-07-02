package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProjectActivity struct {
	ID                 uuid.UUID              `json:"id" db:"id"`
	ProjectID          uuid.UUID              `json:"project_id" db:"project_id"`
	ActivityType       string                 `json:"activity_type" db:"activity_type"` // created, updated, status_changed, member_added, member_removed, milestone_created, milestone_completed, document_uploaded, kpi_updated
	Description        string                 `json:"description" db:"description"`
	Metadata           map[string]interface{} `json:"metadata" db:"metadata"`
	UserID             uuid.UUID              `json:"user_id" db:"user_id"`
	RelatedObjectID    *uuid.UUID             `json:"related_object_id" db:"related_object_id"`
	RelatedObjectType  *string                `json:"related_object_type" db:"related_object_type"`
	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
}
