package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
)

func (r *DB) CreateProject(project *domain.Project) (*domain.Project, error) {
	fmt.Println(project.StartDate, project.EndDate)
	query := `
		INSERT INTO projects (
			id, title, description, domain, status, start_date, end_date,
			progress_percentage, isPublic, owner_id, created_by, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
		RETURNING id, title, description, domain, status, start_date, end_date,
				 progress_percentage, isPublic, owner_id, created_at, updated_at,
				 created_by, updated_by`

	// Generate UUID if not provided
	if project.ID == uuid.Nil {
		project.ID = uuid.New()
	}

	var createdProject domain.Project
	err := r.db.QueryRow(
		query,
		project.ID,
		project.Title,
		project.Description,
		project.Domain,
		project.Status,
		project.StartDate,
		project.EndDate,
		project.ProgressPercentage,
		project.IsPublic,
		project.OwnerID,
		project.CreatedBy,
		project.UpdatedBy,
	).Scan(
		&createdProject.ID,
		&createdProject.Title,
		&createdProject.Description,
		&createdProject.Domain,
		&createdProject.Status,
		&createdProject.StartDate,
		&createdProject.EndDate,
		&createdProject.ProgressPercentage,
		&createdProject.IsPublic,
		&createdProject.OwnerID,
		&createdProject.CreatedAt,
		&createdProject.UpdatedAt,
		&createdProject.CreatedBy,
		&createdProject.UpdatedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return &createdProject, nil
}

func (r *DB) UpdateProject(project *domain.Project) (*domain.Project, error) {
	query := `
		UPDATE projects SET
			title = $2,
			description = $3,
			domain = $4,
			status = $5,
			start_date = $6,
			end_date = $7,
			progress_percentage = $8,
			isPublic = $9,
			updated_by = $10,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, title, description, domain, status, start_date, end_date,
				 progress_percentage, isPublic, owner_id, created_at, updated_at,
				 created_by, updated_by`

	var updatedProject domain.Project
	err := r.db.QueryRow(
		query,
		project.ID,
		project.Title,
		project.Description,
		project.Domain,
		project.Status,
		project.StartDate,
		project.EndDate,
		project.ProgressPercentage,
		project.IsPublic,
		project.UpdatedBy,
	).Scan(
		&updatedProject.ID,
		&updatedProject.Title,
		&updatedProject.Description,
		&updatedProject.Domain,
		&updatedProject.Status,
		&updatedProject.StartDate,
		&updatedProject.EndDate,
		&updatedProject.ProgressPercentage,
		&updatedProject.IsPublic,
		&updatedProject.OwnerID,
		&updatedProject.CreatedAt,
		&updatedProject.UpdatedAt,
		&updatedProject.CreatedBy,
		&updatedProject.UpdatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project with ID %s not found", project.ID)
		}
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return &updatedProject, nil
}

func (r *DB) DeleteProject(projectID uuid.UUID) error {
	query := `DELETE FROM projects WHERE id = $1`

	result, err := r.db.Exec(query, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project with ID %s not found", projectID)
	}

	return nil
}

func (r *DB) GetProjectByID(projectID uuid.UUID) (*domain.Project, error) {
	query := `
		SELECT id, title, description, domain, status, start_date, end_date,
			   progress_percentage, isPublic, owner_id, created_at, updated_at,
			   created_by, updated_by
		FROM projects
		WHERE id = $1`

	var project domain.Project
	err := r.db.QueryRow(query, projectID).Scan(
		&project.ID,
		&project.Title,
		&project.Description,
		&project.Domain,
		&project.Status,
		&project.StartDate,
		&project.EndDate,
		&project.ProgressPercentage,
		&project.IsPublic,
		&project.OwnerID,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.CreatedBy,
		&project.UpdatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project with ID %s not found", projectID)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

func (r *DB) GetAllProjects(limit, offset int, filters map[string]interface{}) ([]*domain.Project, int, error) {
	// Build the WHERE clause based on filters
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	for key, value := range filters {
		switch key {
		case "status":
			whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "domain":
			whereConditions = append(whereConditions, fmt.Sprintf("domain ILIKE $%d", argIndex))
			args = append(args, fmt.Sprintf("%%%s%%", value))
			argIndex++
		case "owner_id":
			whereConditions = append(whereConditions, fmt.Sprintf("owner_id = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "is_public":
			whereConditions = append(whereConditions, fmt.Sprintf("isPublic = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "search":
			whereConditions = append(whereConditions, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM projects %s", whereClause)
	var totalCount int
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	// Get projects with pagination
	query := fmt.Sprintf(`
		SELECT id, title, description, domain, status, start_date, end_date,
			   progress_percentage, isPublic, owner_id, created_at, updated_at,
			   created_by, updated_by
		FROM projects
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		var project domain.Project
		err := rows.Scan(
			&project.ID,
			&project.Title,
			&project.Description,
			&project.Domain,
			&project.Status,
			&project.StartDate,
			&project.EndDate,
			&project.ProgressPercentage,
			&project.IsPublic,
			&project.OwnerID,
			&project.CreatedAt,
			&project.UpdatedAt,
			&project.CreatedBy,
			&project.UpdatedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, &project)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return projects, totalCount, nil
}