package domain

import (
	"time"

	"github.com/google/uuid"
)

type KPI struct {
	ID           uuid.UUID `json:"id" db:"id"`
	ProjectID    uuid.UUID `json:"project_id" db:"project_id"`
	Name         string    `json:"name" db:"name"`
	Description  *string   `json:"description" db:"description"`
	MetricType   string    `json:"metric_type" db:"metric_type"` // percentage, count, currency, hours, days
	LastUpdated  time.Time `json:"last_updated" db:"last_updated"`
	TargetValue  float64   `json:"target_value" db:"target_value"`
	CurrentValue float64   `json:"current_value" db:"current_value"`
	Unit         *string   `json:"unit" db:"unit"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy    uuid.UUID `json:"created_by" db:"created_by"`
}
