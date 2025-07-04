package services

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
)

type projectDocumentService struct {
	documentRepo    ports.ProjectDocumentRepo
	projectRepo     ports.ProjectRepo
	projectMemberRepo ports.ProjectMemberRepo
}

// NewProjectDocumentService creates a new project document service
func NewProjectDocumentService(documentRepo ports.ProjectDocumentRepo, projectRepo ports.ProjectRepo, projectMemberRepo ports.ProjectMemberRepo) ports.ProjectDocumentService {
	return &projectDocumentService{
		documentRepo:    documentRepo,
		projectRepo:     projectRepo,
		projectMemberRepo: projectMemberRepo,
	}
}

// Upload uploads a new document for a project
func (s *projectDocumentService) Upload(ctx context.Context, document *domain.ProjectDocument) (*domain.ProjectDocument, error) {
	// Validate the document
	if err := s.ValidateDocument(ctx, document); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if project exists
	project, err := s.projectRepo.GetProjectByID(document.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project not found")
	}

	// Check if user has permission to upload documents
	canUpload, err := s.canUploadDocument(ctx, document.ProjectID, document.UploadedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to check upload permissions: %w", err)
	}
	if !canUpload {
		return nil, errors.New("user does not have permission to upload documents to this project")
	}

	// Set default version if not provided
	if document.Version == "" {
		document.Version = "1.0"
	}

	// Set default active status
	document.IsActive = true

	// Create the document
	createdDocument, err := s.documentRepo.CreateDocument(document)
	if err != nil {
		return nil, fmt.Errorf("failed to upload document: %w", err)
	}

	return createdDocument, nil
}

// Update updates an existing document
func (s *projectDocumentService) Update(ctx context.Context, document *domain.ProjectDocument) (*domain.ProjectDocument, error) {
	// Check if document exists
	existingDocument, err := s.documentRepo.GetDocumentByID(document.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	if existingDocument == nil {
		return nil, errors.New("document not found")
	}

	// Check if document can be modified
	canModify, err := s.CanModifyDocument(ctx, document.ID, document.UploadedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to check modification permissions: %w", err)
	}
	if !canModify {
		return nil, errors.New("document cannot be modified")
	}

	// Validate the updated document
	if err := s.ValidateDocument(ctx, document); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Preserve immutable fields
	document.ProjectID = existingDocument.ProjectID
	document.UploadedBy = existingDocument.UploadedBy

	// Update the document
	updatedDocument, err := s.documentRepo.UpdateDocument(document)
	if err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	return updatedDocument, nil
}

// Delete deletes a document by ID
func (s *projectDocumentService) Delete(ctx context.Context, documentID uuid.UUID) error {
	// Check if document exists
	document, err := s.documentRepo.GetDocumentByID(documentID)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}
	if document == nil {
		return errors.New("document not found")
	}

	// Check if document can be modified
	canModify, err := s.CanModifyDocument(ctx, documentID, document.UploadedBy)
	if err != nil {
		return fmt.Errorf("failed to check modification permissions: %w", err)
	}
	if !canModify {
		return errors.New("document cannot be deleted")
	}

	// Delete the document
	if err := s.documentRepo.DeleteDocument(documentID); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// GetByID retrieves a document by its ID
func (s *projectDocumentService) GetByID(ctx context.Context, documentID uuid.UUID) (*domain.ProjectDocument, error) {
	document, err := s.documentRepo.GetDocumentByID(documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	return document, nil
}

// GetProjectDocuments retrieves all documents for a specific project
func (s *projectDocumentService) GetProjectDocuments(ctx context.Context, projectID uuid.UUID, filters map[string]interface{}) ([]*domain.ProjectDocument, error) {
	// Check if project exists
	project, err := s.projectRepo.GetProjectByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project not found")
	}

	documents, err := s.documentRepo.GetProjectDocuments(projectID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get project documents: %w", err)
	}
	return documents, nil
}

// GetAllDocuments retrieves all documents with optional filtering
func (s *projectDocumentService) GetAllDocuments(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*domain.ProjectDocument, int, error) {
	// Validate pagination parameters
	if limit <= 0 {
		limit = 10 // Default limit
	}
	if limit > 100 {
		limit = 100 // Maximum limit
	}
	if offset < 0 {
		offset = 0
	}

	documents, total, err := s.documentRepo.GetAllDocuments(limit, offset, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get documents: %w", err)
	}
	return documents, total, nil
}

// Deactivate deactivates a document instead of deleting it
func (s *projectDocumentService) Deactivate(ctx context.Context, documentID uuid.UUID) (*domain.ProjectDocument, error) {
	// Check if document exists
	document, err := s.documentRepo.GetDocumentByID(documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	if document == nil {
		return nil, errors.New("document not found")
	}

	// Check if document can be modified
	canModify, err := s.CanModifyDocument(ctx, documentID, document.UploadedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to check modification permissions: %w", err)
	}
	if !canModify {
		return nil, errors.New("document cannot be deactivated")
	}

	// Deactivate the document
	deactivatedDocument, err := s.documentRepo.DeactivateDocument(documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to deactivate document: %w", err)
	}

	return deactivatedDocument, nil
}

// GetDocumentsByType retrieves documents by type
func (s *projectDocumentService) GetDocumentsByType(ctx context.Context, documentType string, projectID *uuid.UUID) ([]*domain.ProjectDocument, error) {
	// Validate document type
	validDocumentTypes := map[string]bool{
		"requirement": true,
		"design":      true,
		"technical":   true,
		"contract":    true,
		"report":      true,
		"other":       true,
	}
	if !validDocumentTypes[documentType] {
		return nil, errors.New("invalid document type")
	}

	documents, err := s.documentRepo.GetDocumentsByType(documentType, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents by type: %w", err)
	}
	return documents, nil
}

// GetDocumentsByUploader retrieves documents by uploader
func (s *projectDocumentService) GetDocumentsByUploader(ctx context.Context, uploaderID uuid.UUID, projectID *uuid.UUID) ([]*domain.ProjectDocument, error) {
	documents, err := s.documentRepo.GetDocumentsByUploader(uploaderID, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents by uploader: %w", err)
	}
	return documents, nil
}

// ValidateDocument validates document data and business rules
func (s *projectDocumentService) ValidateDocument(ctx context.Context, document *domain.ProjectDocument) error {
	if document == nil {
		return errors.New("document cannot be nil")
	}

	// Validate required fields
	if document.ProjectID == uuid.Nil {
		return errors.New("project ID is required")
	}

	if document.Name == "" {
		return errors.New("document name is required")
	}

	if len(document.Name) > 255 {
		return errors.New("document name cannot exceed 255 characters")
	}

	if document.Description != nil && len(*document.Description) > 2000 {
		return errors.New("document description cannot exceed 2000 characters")
	}

	if document.FileURL == "" {
		return errors.New("file URL is required")
	}

	if document.UploadedBy == uuid.Nil {
		return errors.New("uploader ID is required")
	}

	// Validate document type
	validDocumentTypes := map[string]bool{
		"requirement": true,
		"design":      true,
		"technical":   true,
		"contract":    true,
		"report":      true,
		"other":       true,
	}
	if document.DocumentType == "" {
		return errors.New("document type is required")
	}
	if !validDocumentTypes[document.DocumentType] {
		return errors.New("invalid document type")
	}

	// Validate file size
	if document.FileSize != nil && *document.FileSize < 0 {
		return errors.New("file size cannot be negative")
	}

	if document.FileSize != nil && *document.FileSize > 100*1024*1024 { // 100MB limit
		return errors.New("file size cannot exceed 100MB")
	}

	// Validate version format
	if document.Version != "" && len(document.Version) > 20 {
		return errors.New("version string cannot exceed 20 characters")
	}

	// Validate MIME type
	if document.MimeType != nil && len(*document.MimeType) > 100 {
		return errors.New("MIME type cannot exceed 100 characters")
	}

	// Validate file extension matches MIME type (basic check)
	if document.MimeType != nil && document.FileURL != "" {
		ext := strings.ToLower(filepath.Ext(document.FileURL))
		mimeType := *document.MimeType
		
		// Basic MIME type validation
		validExtensions := map[string][]string{
			"application/pdf":  {".pdf"},
			"text/plain":       {".txt"},
			"application/msword": {".doc"},
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {".docx"},
			"image/jpeg":       {".jpg", ".jpeg"},
			"image/png":        {".png"},
			"image/gif":        {".gif"},
		}
		
		if exts, exists := validExtensions[mimeType]; exists {
			isValidExt := false
			for _, validExtension := range exts {
				if ext == validExtension {
					isValidExt = true
					break
				}
			}
			if !isValidExt {
				return fmt.Errorf("file extension %s does not match MIME type %s", ext, mimeType)
			}
		}
	}

	return nil
}

// CanModifyDocument checks if a document can be modified
func (s *projectDocumentService) CanModifyDocument(ctx context.Context, documentID uuid.UUID, userID uuid.UUID) (bool, error) {
	document, err := s.documentRepo.GetDocumentByID(documentID)
	if err != nil {
		return false, fmt.Errorf("failed to get document: %w", err)
	}
	if document == nil {
		return false, errors.New("document not found")
	}

	// Check if the project is still active
	project, err := s.projectRepo.GetProjectByID(document.ProjectID)
	if err != nil {
		return false, fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return false, errors.New("project not found")
	}

	// Check project status
	if project.Status == "completed" || project.Status == "cancelled" {
		return false, nil
	}

	// Check if user is the uploader
	if document.UploadedBy == userID {
		return true, nil
	}

	// Check if user is project owner
	if project.OwnerID == userID {
		return true, nil
	}

	// Check if user has edit permissions
	canUpload, err := s.canUploadDocument(ctx, document.ProjectID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check user permissions: %w", err)
	}

	return canUpload, nil
}

// GetDocumentStatistics returns statistics for documents
func (s *projectDocumentService) GetDocumentStatistics(ctx context.Context, projectID *uuid.UUID) (*ports.DocumentStatistics, error) {
	stats, err := s.documentRepo.GetDocumentStatistics(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document statistics: %w", err)
	}
	return stats, nil
}

// GetLatestVersion gets the latest version of a document by name
func (s *projectDocumentService) GetLatestVersion(ctx context.Context, projectID uuid.UUID, documentName string) (*domain.ProjectDocument, error) {
	// Check if project exists
	project, err := s.projectRepo.GetProjectByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return nil, errors.New("project not found")
	}

	document, err := s.documentRepo.GetLatestVersion(projectID, documentName)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}
	return document, nil
}

// CreateNewVersion creates a new version of an existing document
func (s *projectDocumentService) CreateNewVersion(ctx context.Context, document *domain.ProjectDocument, baseDocumentID uuid.UUID) (*domain.ProjectDocument, error) {
	// Get the base document
	baseDocument, err := s.documentRepo.GetDocumentByID(baseDocumentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get base document: %w", err)
	}
	if baseDocument == nil {
		return nil, errors.New("base document not found")
	}

	// Check permissions
	canModify, err := s.CanModifyDocument(ctx, baseDocumentID, document.UploadedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to check modification permissions: %w", err)
	}
	if !canModify {
		return nil, errors.New("user does not have permission to create new version of this document")
	}

	// Copy relevant fields from base document
	document.ProjectID = baseDocument.ProjectID
	document.Name = baseDocument.Name
	document.DocumentType = baseDocument.DocumentType

	// Generate new version number
	newVersion, err := s.generateNewVersion(baseDocument.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new version: %w", err)
	}
	document.Version = newVersion

	// Validate and create the new version
	return s.Upload(ctx, document)
}

// GetDocumentAnalysis returns detailed analysis of project documents
func (s *projectDocumentService) GetDocumentAnalysis(ctx context.Context, projectID *uuid.UUID) (*ports.DocumentAnalysis, error) {
	// Get all documents for analysis
	var allDocuments []*domain.ProjectDocument
	var err error

	if projectID != nil {
		allDocuments, err = s.documentRepo.GetProjectDocuments(*projectID, nil)
	} else {
		allDocuments, _, err = s.documentRepo.GetAllDocuments(1000, 0, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get documents for analysis: %w", err)
	}

	analysis := &ports.DocumentAnalysis{
		DocumentsByType:       make(map[string][]*domain.ProjectDocument),
		LargestDocuments:      []*domain.ProjectDocument{},
		RecentDocuments:       []*domain.ProjectDocument{},
		MostActiveUploaders:   []ports.UploaderStats{},
		DocumentTypeAnalysis:  make(map[string]ports.TypeAnalysis),
		StorageRecommendations: []string{},
	}

	if len(allDocuments) == 0 {
		return analysis, nil
	}

	// Group documents by type
	uploaderStats := make(map[string]*ports.UploaderStats)
	typeStats := make(map[string]*ports.TypeAnalysis)

	for _, doc := range allDocuments {
		// Group by type
		analysis.DocumentsByType[doc.DocumentType] = append(analysis.DocumentsByType[doc.DocumentType], doc)

		// Track uploader stats
		uploaderKey := doc.UploadedBy.String()
		if _, exists := uploaderStats[uploaderKey]; !exists {
			uploaderStats[uploaderKey] = &ports.UploaderStats{
				UploaderID: uploaderKey,
			}
		}
		uploaderStats[uploaderKey].DocumentCount++
		if doc.FileSize != nil {
			uploaderStats[uploaderKey].TotalSize += *doc.FileSize
		}

		// Track type stats
		if _, exists := typeStats[doc.DocumentType]; !exists {
			typeStats[doc.DocumentType] = &ports.TypeAnalysis{}
		}
		typeStats[doc.DocumentType].Count++
		if doc.FileSize != nil {
			typeStats[doc.DocumentType].TotalSize += *doc.FileSize
		}
	}

	// Calculate type averages
	for docType, stats := range typeStats {
		if stats.Count > 0 {
			stats.AverageSize = float64(stats.TotalSize) / float64(stats.Count)
		}
		analysis.DocumentTypeAnalysis[docType] = *stats
	}

	// Convert uploader stats to slice and sort
	for _, stats := range uploaderStats {
		analysis.MostActiveUploaders = append(analysis.MostActiveUploaders, *stats)
	}
	sort.Slice(analysis.MostActiveUploaders, func(i, j int) bool {
		return analysis.MostActiveUploaders[i].DocumentCount > analysis.MostActiveUploaders[j].DocumentCount
	})

	// Get largest documents
	documentsWithSize := make([]*domain.ProjectDocument, 0)
	for _, doc := range allDocuments {
		if doc.FileSize != nil && *doc.FileSize > 0 {
			documentsWithSize = append(documentsWithSize, doc)
		}
	}
	sort.Slice(documentsWithSize, func(i, j int) bool {
		return *documentsWithSize[i].FileSize > *documentsWithSize[j].FileSize
	})
	if len(documentsWithSize) > 10 {
		analysis.LargestDocuments = documentsWithSize[:10]
	} else {
		analysis.LargestDocuments = documentsWithSize
	}

	// Get recent documents (last 30 days)
	// This would be better with actual date filtering, but for simplicity, we'll take the most recent by creation
	recentDocs := make([]*domain.ProjectDocument, len(allDocuments))
	copy(recentDocs, allDocuments)
	sort.Slice(recentDocs, func(i, j int) bool {
		return recentDocs[i].CreatedAt.After(recentDocs[j].CreatedAt)
	})
	if len(recentDocs) > 10 {
		analysis.RecentDocuments = recentDocs[:10]
	} else {
		analysis.RecentDocuments = recentDocs
	}

	// Generate storage recommendations
	totalSize := int64(0)
	for _, doc := range allDocuments {
		if doc.FileSize != nil {
			totalSize += *doc.FileSize
		}
	}

	if totalSize > 500*1024*1024 { // 500MB
		analysis.StorageRecommendations = append(analysis.StorageRecommendations, "Consider archiving old document versions to reduce storage usage.")
	}

	if len(analysis.DocumentsByType["other"]) > len(allDocuments)/4 {
		analysis.StorageRecommendations = append(analysis.StorageRecommendations, "Many documents are classified as 'other'. Consider adding more specific document types.")
	}

	if len(analysis.MostActiveUploaders) > 0 && analysis.MostActiveUploaders[0].DocumentCount > 50 {
		analysis.StorageRecommendations = append(analysis.StorageRecommendations, "Some users have uploaded many documents. Consider implementing document organization guidelines.")
	}

	return analysis, nil
}

// Helper methods

func (s *projectDocumentService) canUploadDocument(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) (bool, error) {
	// Check if user is project member with appropriate permissions
	member, err := s.projectMemberRepo.GetUserProjectMembership(userID, projectID)
	if err != nil {
		return false, fmt.Errorf("failed to check project membership: %w", err)
	}
	if member == nil {
		return false, nil // User is not a member
	}

	// Check if member is active and has permission to upload documents
	if !member.IsActive {
		return false, nil
	}

	// Members with edit or manage permissions can upload documents
	return member.CanEditProject || member.CanManageTasks, nil
}

func (s *projectDocumentService) generateNewVersion(currentVersion string) (string, error) {
	// Simple version increment logic
	// This assumes versions are in format "X.Y" or just "X"
	
	if currentVersion == "" {
		return "1.0", nil
	}

	parts := strings.Split(currentVersion, ".")
	if len(parts) == 1 {
		// Single number version
		if major, err := strconv.Atoi(parts[0]); err == nil {
			return fmt.Sprintf("%d.0", major+1), nil
		}
		return "2.0", nil
	}

	if len(parts) == 2 {
		// Major.Minor version
		major, err1 := strconv.Atoi(parts[0])
		minor, err2 := strconv.Atoi(parts[1])
		
		if err1 == nil && err2 == nil {
			return fmt.Sprintf("%d.%d", major, minor+1), nil
		}
	}

	// Fallback: append ".1" to current version
	return currentVersion + ".1", nil
}
