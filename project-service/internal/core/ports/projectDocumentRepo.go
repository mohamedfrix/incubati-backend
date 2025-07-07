package ports

import (
	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
)

type ProjectDocumentRepo interface {
	CreateDocument(document *domain.ProjectDocument) (*domain.ProjectDocument, error)
	UpdateDocument(document *domain.ProjectDocument) (*domain.ProjectDocument, error)
	DeleteDocument(documentID uuid.UUID) error
	GetDocumentByID(documentID uuid.UUID) (*domain.ProjectDocument, error)
	GetProjectDocuments(projectID uuid.UUID, filters map[string]interface{}) ([]*domain.ProjectDocument, error)
	GetAllDocuments(limit, offset int, filters map[string]interface{}) ([]*domain.ProjectDocument, int, error)
	DeactivateDocument(documentID uuid.UUID) (*domain.ProjectDocument, error)
	GetDocumentsByType(documentType string, projectID *uuid.UUID) ([]*domain.ProjectDocument, error)
	GetDocumentsByUploader(uploaderID uuid.UUID, projectID *uuid.UUID) ([]*domain.ProjectDocument, error)
	GetDocumentStatistics(projectID *uuid.UUID) (*DocumentStatistics, error)
	GetLatestVersion(projectID uuid.UUID, documentName string) (*domain.ProjectDocument, error)
}

// DocumentStatistics represents document analytics
type DocumentStatistics struct {
	TotalDocuments       int                    `json:"total_documents"`
	ActiveDocuments      int                    `json:"active_documents"`
	DocumentsByType      map[string]int         `json:"documents_by_type"`
	TotalFileSize        int64                  `json:"total_file_size"`
	AverageFileSize      float64                `json:"average_file_size"`
	RecentDocuments      int                    `json:"recent_documents_30_days"`
	DocumentsByUploader  map[string]int         `json:"documents_by_uploader"`
}
