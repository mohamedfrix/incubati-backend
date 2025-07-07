package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProjectDocument struct {
	ID           uuid.UUID `json:"id" db:"id"`
	ProjectID    uuid.UUID `json:"project_id" db:"project_id"`
	Name         string    `json:"name" db:"name"`
	Description  *string   `json:"description" db:"description"`
	DocumentType string    `json:"document_type" db:"document_type"` // requirement, design, technical, contract, report, other
	FileURL      string    `json:"file_url" db:"file_url"`
	FileSize     *int64    `json:"file_size" db:"file_size"`
	MimeType     *string   `json:"mime_type" db:"mime_type"`
	Version      string    `json:"version" db:"version"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	UploadedBy   uuid.UUID `json:"uploaded_by" db:"uploaded_by"`
}
