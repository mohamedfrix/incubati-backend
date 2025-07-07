package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
)

// ProjectDocumentService defines the interface for project document business logic
type ProjectDocumentService interface {
	// Upload uploads a new document for a project
	Upload(ctx context.Context, document *domain.ProjectDocument) (*domain.ProjectDocument, error)
	
	// Update updates an existing document
	Update(ctx context.Context, document *domain.ProjectDocument) (*domain.ProjectDocument, error)
	
	// Delete deletes a document by ID
	Delete(ctx context.Context, documentID uuid.UUID) error
	
	// GetByID retrieves a document by its ID
	GetByID(ctx context.Context, documentID uuid.UUID) (*domain.ProjectDocument, error)
	
	// GetProjectDocuments retrieves all documents for a specific project
	GetProjectDocuments(ctx context.Context, projectID uuid.UUID, filters map[string]interface{}) ([]*domain.ProjectDocument, error)
	
	// GetAllDocuments retrieves all documents with optional filtering
	GetAllDocuments(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*domain.ProjectDocument, int, error)
	
	// Deactivate deactivates a document instead of deleting it
	Deactivate(ctx context.Context, documentID uuid.UUID) (*domain.ProjectDocument, error)
	
	// GetDocumentsByType retrieves documents by type
	GetDocumentsByType(ctx context.Context, documentType string, projectID *uuid.UUID) ([]*domain.ProjectDocument, error)
	
	// GetDocumentsByUploader retrieves documents by uploader
	GetDocumentsByUploader(ctx context.Context, uploaderID uuid.UUID, projectID *uuid.UUID) ([]*domain.ProjectDocument, error)
	
	// ValidateDocument validates document data and business rules
	ValidateDocument(ctx context.Context, document *domain.ProjectDocument) error
	
	// CanModifyDocument checks if a document can be modified
	CanModifyDocument(ctx context.Context, documentID uuid.UUID, userID uuid.UUID) (bool, error)
	
	// GetDocumentStatistics returns statistics for documents
	GetDocumentStatistics(ctx context.Context, projectID *uuid.UUID) (*DocumentStatistics, error)
	
	// GetLatestVersion gets the latest version of a document by name
	GetLatestVersion(ctx context.Context, projectID uuid.UUID, documentName string) (*domain.ProjectDocument, error)
	
	// CreateNewVersion creates a new version of an existing document
	CreateNewVersion(ctx context.Context, document *domain.ProjectDocument, baseDocumentID uuid.UUID) (*domain.ProjectDocument, error)
	
	// GetDocumentAnalysis returns detailed analysis of project documents
	GetDocumentAnalysis(ctx context.Context, projectID *uuid.UUID) (*DocumentAnalysis, error)
}

// DocumentAnalysis represents detailed document analysis
type DocumentAnalysis struct {
	DocumentsByType       map[string][]*domain.ProjectDocument `json:"documents_by_type"`
	LargestDocuments      []*domain.ProjectDocument            `json:"largest_documents"`
	RecentDocuments       []*domain.ProjectDocument            `json:"recent_documents"`
	MostActiveUploaders   []UploaderStats                      `json:"most_active_uploaders"`
	DocumentTypeAnalysis  map[string]TypeAnalysis              `json:"document_type_analysis"`
	StorageRecommendations []string                            `json:"storage_recommendations"`
}

// UploaderStats represents uploader statistics
type UploaderStats struct {
	UploaderID    string `json:"uploader_id"`
	DocumentCount int    `json:"document_count"`
	TotalSize     int64  `json:"total_size"`
}

// TypeAnalysis represents analysis per document type
type TypeAnalysis struct {
	Count       int     `json:"count"`
	TotalSize   int64   `json:"total_size"`
	AverageSize float64 `json:"average_size"`
}
