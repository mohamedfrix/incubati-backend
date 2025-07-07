package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
)

func (r *DB) CreateDocument(document *domain.ProjectDocument) (*domain.ProjectDocument, error) {
	query := `
		INSERT INTO project_documents (
			id, project_id, name, description, document_type, file_url,
			file_size, mime_type, version, is_active, uploaded_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
		RETURNING id, project_id, name, description, document_type, file_url,
				 file_size, mime_type, version, is_active, created_at, updated_at, uploaded_by`

	// Generate UUID if not provided
	if document.ID == uuid.Nil {
		document.ID = uuid.New()
	}

	var createdDocument domain.ProjectDocument
	err := r.db.QueryRow(
		query,
		document.ID,
		document.ProjectID,
		document.Name,
		document.Description,
		document.DocumentType,
		document.FileURL,
		document.FileSize,
		document.MimeType,
		document.Version,
		document.IsActive,
		document.UploadedBy,
	).Scan(
		&createdDocument.ID,
		&createdDocument.ProjectID,
		&createdDocument.Name,
		&createdDocument.Description,
		&createdDocument.DocumentType,
		&createdDocument.FileURL,
		&createdDocument.FileSize,
		&createdDocument.MimeType,
		&createdDocument.Version,
		&createdDocument.IsActive,
		&createdDocument.CreatedAt,
		&createdDocument.UpdatedAt,
		&createdDocument.UploadedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return &createdDocument, nil
}

func (r *DB) UpdateDocument(document *domain.ProjectDocument) (*domain.ProjectDocument, error) {
	query := `
		UPDATE project_documents SET
			name = $2,
			description = $3,
			document_type = $4,
			file_url = $5,
			file_size = $6,
			mime_type = $7,
			version = $8,
			is_active = $9,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, project_id, name, description, document_type, file_url,
				 file_size, mime_type, version, is_active, created_at, updated_at, uploaded_by`

	var updatedDocument domain.ProjectDocument
	err := r.db.QueryRow(
		query,
		document.ID,
		document.Name,
		document.Description,
		document.DocumentType,
		document.FileURL,
		document.FileSize,
		document.MimeType,
		document.Version,
		document.IsActive,
	).Scan(
		&updatedDocument.ID,
		&updatedDocument.ProjectID,
		&updatedDocument.Name,
		&updatedDocument.Description,
		&updatedDocument.DocumentType,
		&updatedDocument.FileURL,
		&updatedDocument.FileSize,
		&updatedDocument.MimeType,
		&updatedDocument.Version,
		&updatedDocument.IsActive,
		&updatedDocument.CreatedAt,
		&updatedDocument.UpdatedAt,
		&updatedDocument.UploadedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document with ID %s not found", document.ID)
		}
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	return &updatedDocument, nil
}

func (r *DB) DeleteDocument(documentID uuid.UUID) error {
	// Use DELETE with RETURNING to get document info in one query
	query := `DELETE FROM project_documents WHERE id = $1 RETURNING name`
	
	var documentName string
	err := r.db.QueryRow(query, documentID).Scan(&documentName)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("document with ID %s not found", documentID)
		}
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

func (r *DB) GetDocumentByID(documentID uuid.UUID) (*domain.ProjectDocument, error) {
	query := `
		SELECT id, project_id, name, description, document_type, file_url,
			   file_size, mime_type, version, is_active, created_at, updated_at, uploaded_by
		FROM project_documents
		WHERE id = $1`

	var document domain.ProjectDocument
	err := r.db.QueryRow(query, documentID).Scan(
		&document.ID,
		&document.ProjectID,
		&document.Name,
		&document.Description,
		&document.DocumentType,
		&document.FileURL,
		&document.FileSize,
		&document.MimeType,
		&document.Version,
		&document.IsActive,
		&document.CreatedAt,
		&document.UpdatedAt,
		&document.UploadedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document with ID %s not found", documentID)
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return &document, nil
}

func (r *DB) GetProjectDocuments(projectID uuid.UUID, filters map[string]interface{}) ([]*domain.ProjectDocument, error) {
	// Build the WHERE clause based on filters
	whereConditions := []string{"project_id = $1"}
	args := []interface{}{projectID}
	argIndex := 2

	for key, value := range filters {
		switch key {
		case "document_type":
			whereConditions = append(whereConditions, fmt.Sprintf("document_type = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "is_active":
			whereConditions = append(whereConditions, fmt.Sprintf("is_active = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "uploaded_by":
			whereConditions = append(whereConditions, fmt.Sprintf("uploaded_by = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "mime_type":
			whereConditions = append(whereConditions, fmt.Sprintf("mime_type = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "version":
			whereConditions = append(whereConditions, fmt.Sprintf("version = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "recent_days":
			days := value.(int)
			whereConditions = append(whereConditions, fmt.Sprintf("created_at >= CURRENT_DATE - INTERVAL '%d days'", days))
		}
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := fmt.Sprintf(`
		SELECT id, project_id, name, description, document_type, file_url,
			   file_size, mime_type, version, is_active, created_at, updated_at, uploaded_by
		FROM project_documents
		WHERE %s
		ORDER BY created_at DESC`, whereClause)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query project documents: %w", err)
	}
	defer rows.Close()

	var documents []*domain.ProjectDocument
	for rows.Next() {
		var document domain.ProjectDocument
		err := rows.Scan(
			&document.ID,
			&document.ProjectID,
			&document.Name,
			&document.Description,
			&document.DocumentType,
			&document.FileURL,
			&document.FileSize,
			&document.MimeType,
			&document.Version,
			&document.IsActive,
			&document.CreatedAt,
			&document.UpdatedAt,
			&document.UploadedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		documents = append(documents, &document)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return documents, nil
}

func (r *DB) GetAllDocuments(limit, offset int, filters map[string]interface{}) ([]*domain.ProjectDocument, int, error) {
	// Build the WHERE clause based on filters
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	for key, value := range filters {
		switch key {
		case "project_id":
			whereConditions = append(whereConditions, fmt.Sprintf("project_id = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "document_type":
			whereConditions = append(whereConditions, fmt.Sprintf("document_type = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "is_active":
			whereConditions = append(whereConditions, fmt.Sprintf("is_active = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "uploaded_by":
			whereConditions = append(whereConditions, fmt.Sprintf("uploaded_by = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "search":
			whereConditions = append(whereConditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
			searchTerm := fmt.Sprintf("%%%s%%", value)
			args = append(args, searchTerm)
			argIndex++
		}
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM project_documents %s", whereClause)
	var totalCount int
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Get documents with pagination
	query := fmt.Sprintf(`
		SELECT id, project_id, name, description, document_type, file_url,
			   file_size, mime_type, version, is_active, created_at, updated_at, uploaded_by
		FROM project_documents
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query documents: %w", err)
	}
	defer rows.Close()

	var documents []*domain.ProjectDocument
	for rows.Next() {
		var document domain.ProjectDocument
		err := rows.Scan(
			&document.ID,
			&document.ProjectID,
			&document.Name,
			&document.Description,
			&document.DocumentType,
			&document.FileURL,
			&document.FileSize,
			&document.MimeType,
			&document.Version,
			&document.IsActive,
			&document.CreatedAt,
			&document.UpdatedAt,
			&document.UploadedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan document: %w", err)
		}
		documents = append(documents, &document)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return documents, totalCount, nil
}

func (r *DB) DeactivateDocument(documentID uuid.UUID) (*domain.ProjectDocument, error) {
	query := `
		UPDATE project_documents SET
			is_active = false,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, project_id, name, description, document_type, file_url,
				 file_size, mime_type, version, is_active, created_at, updated_at, uploaded_by`

	var document domain.ProjectDocument
	err := r.db.QueryRow(query, documentID).Scan(
		&document.ID,
		&document.ProjectID,
		&document.Name,
		&document.Description,
		&document.DocumentType,
		&document.FileURL,
		&document.FileSize,
		&document.MimeType,
		&document.Version,
		&document.IsActive,
		&document.CreatedAt,
		&document.UpdatedAt,
		&document.UploadedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document with ID %s not found", documentID)
		}
		return nil, fmt.Errorf("failed to deactivate document: %w", err)
	}

	return &document, nil
}

func (r *DB) GetDocumentsByType(documentType string, projectID *uuid.UUID) ([]*domain.ProjectDocument, error) {
	query := `
		SELECT id, project_id, name, description, document_type, file_url,
			   file_size, mime_type, version, is_active, created_at, updated_at, uploaded_by
		FROM project_documents
		WHERE document_type = $1`
	
	args := []interface{}{documentType}
	
	if projectID != nil {
		query += " AND project_id = $2"
		args = append(args, *projectID)
	}
	
	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents by type: %w", err)
	}
	defer rows.Close()

	var documents []*domain.ProjectDocument
	for rows.Next() {
		var document domain.ProjectDocument
		err := rows.Scan(
			&document.ID,
			&document.ProjectID,
			&document.Name,
			&document.Description,
			&document.DocumentType,
			&document.FileURL,
			&document.FileSize,
			&document.MimeType,
			&document.Version,
			&document.IsActive,
			&document.CreatedAt,
			&document.UpdatedAt,
			&document.UploadedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		documents = append(documents, &document)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return documents, nil
}

func (r *DB) GetDocumentsByUploader(uploaderID uuid.UUID, projectID *uuid.UUID) ([]*domain.ProjectDocument, error) {
	query := `
		SELECT id, project_id, name, description, document_type, file_url,
			   file_size, mime_type, version, is_active, created_at, updated_at, uploaded_by
		FROM project_documents
		WHERE uploaded_by = $1`
	
	args := []interface{}{uploaderID}
	
	if projectID != nil {
		query += " AND project_id = $2"
		args = append(args, *projectID)
	}
	
	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents by uploader: %w", err)
	}
	defer rows.Close()

	var documents []*domain.ProjectDocument
	for rows.Next() {
		var document domain.ProjectDocument
		err := rows.Scan(
			&document.ID,
			&document.ProjectID,
			&document.Name,
			&document.Description,
			&document.DocumentType,
			&document.FileURL,
			&document.FileSize,
			&document.MimeType,
			&document.Version,
			&document.IsActive,
			&document.CreatedAt,
			&document.UpdatedAt,
			&document.UploadedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}
		documents = append(documents, &document)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return documents, nil
}

func (r *DB) GetDocumentStatistics(projectID *uuid.UUID) (*ports.DocumentStatistics, error) {
	query := `
		SELECT 
			COUNT(*) as total_documents,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_documents,
			COALESCE(SUM(file_size), 0) as total_file_size,
			COUNT(CASE WHEN created_at >= CURRENT_DATE - INTERVAL '30 days' THEN 1 END) as recent_documents
		FROM project_documents`
	
	args := []interface{}{}
	
	if projectID != nil {
		query += " WHERE project_id = $1"
		args = append(args, *projectID)
	}

	var stats ports.DocumentStatistics
	var totalFileSize sql.NullInt64

	err := r.db.QueryRow(query, args...).Scan(
		&stats.TotalDocuments,
		&stats.ActiveDocuments,
		&totalFileSize,
		&stats.RecentDocuments,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get document statistics: %w", err)
	}

	stats.TotalFileSize = totalFileSize.Int64
	if stats.TotalDocuments > 0 {
		stats.AverageFileSize = float64(stats.TotalFileSize) / float64(stats.TotalDocuments)
	}

	// Get documents by type
	typeQuery := `
		SELECT document_type, COUNT(*) 
		FROM project_documents`
	
	if projectID != nil {
		typeQuery += " WHERE project_id = $1"
	}
	
	typeQuery += " GROUP BY document_type"

	typeRows, err := r.db.Query(typeQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get document type statistics: %w", err)
	}
	defer typeRows.Close()

	stats.DocumentsByType = make(map[string]int)
	for typeRows.Next() {
		var docType string
		var count int
		err := typeRows.Scan(&docType, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan document type: %w", err)
		}
		stats.DocumentsByType[docType] = count
	}

	// Get documents by uploader
	uploaderQuery := `
		SELECT uploaded_by::text, COUNT(*) 
		FROM project_documents`
	
	if projectID != nil {
		uploaderQuery += " WHERE project_id = $1"
	}
	
	uploaderQuery += " GROUP BY uploaded_by"

	uploaderRows, err := r.db.Query(uploaderQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get uploader statistics: %w", err)
	}
	defer uploaderRows.Close()

	stats.DocumentsByUploader = make(map[string]int)
	for uploaderRows.Next() {
		var uploaderID string
		var count int
		err := uploaderRows.Scan(&uploaderID, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan uploader: %w", err)
		}
		stats.DocumentsByUploader[uploaderID] = count
	}

	return &stats, nil
}

func (r *DB) GetLatestVersion(projectID uuid.UUID, documentName string) (*domain.ProjectDocument, error) {
	query := `
		SELECT id, project_id, name, description, document_type, file_url,
			   file_size, mime_type, version, is_active, created_at, updated_at, uploaded_by
		FROM project_documents
		WHERE project_id = $1 AND name = $2 AND is_active = true
		ORDER BY version DESC, created_at DESC
		LIMIT 1`

	var document domain.ProjectDocument
	err := r.db.QueryRow(query, projectID, documentName).Scan(
		&document.ID,
		&document.ProjectID,
		&document.Name,
		&document.Description,
		&document.DocumentType,
		&document.FileURL,
		&document.FileSize,
		&document.MimeType,
		&document.Version,
		&document.IsActive,
		&document.CreatedAt,
		&document.UpdatedAt,
		&document.UploadedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active document found with name %s in project %s", documentName, projectID)
		}
		return nil, fmt.Errorf("failed to get latest document version: %w", err)
	}

	return &document, nil
}
