package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
)

func (r *DB) CreateMilestone(milestone *domain.Milestone) (*domain.Milestone, error) {
	query := `
		INSERT INTO milestones (
			id, project_id, title, description, due_date, completed_date,
			isCompleted, status, progress_percentage, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
		RETURNING id, project_id, title, description, due_date, completed_date,
				 isCompleted, status, progress_percentage, created_at, updated_at, created_by`

	// Generate UUID if not provided
	if milestone.ID == uuid.Nil {
		milestone.ID = uuid.New()
	}

	var createdMilestone domain.Milestone
	err := r.db.QueryRow(
		query,
		milestone.ID,
		milestone.ProjectID,
		milestone.Title,
		milestone.Description,
		milestone.DueDate,
		milestone.CompletedDate,
		milestone.IsCompleted,
		milestone.Status,
		milestone.ProgressPercentage,
		milestone.CreatedBy,
	).Scan(
		&createdMilestone.ID,
		&createdMilestone.ProjectID,
		&createdMilestone.Title,
		&createdMilestone.Description,
		&createdMilestone.DueDate,
		&createdMilestone.CompletedDate,
		&createdMilestone.IsCompleted,
		&createdMilestone.Status,
		&createdMilestone.ProgressPercentage,
		&createdMilestone.CreatedAt,
		&createdMilestone.UpdatedAt,
		&createdMilestone.CreatedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create milestone: %w", err)
	}

	return &createdMilestone, nil
}

func (r *DB) UpdateMilestone(milestone *domain.Milestone) (*domain.Milestone, error) {
	query := `
		UPDATE milestones SET
			title = $2,
			description = $3,
			due_date = $4,
			completed_date = $5,
			isCompleted = $6,
			status = $7,
			progress_percentage = $8,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, project_id, title, description, due_date, completed_date,
				 isCompleted, status, progress_percentage, created_at, updated_at, created_by`

	var updatedMilestone domain.Milestone
	err := r.db.QueryRow(
		query,
		milestone.ID,
		milestone.Title,
		milestone.Description,
		milestone.DueDate,
		milestone.CompletedDate,
		milestone.IsCompleted,
		milestone.Status,
		milestone.ProgressPercentage,
	).Scan(
		&updatedMilestone.ID,
		&updatedMilestone.ProjectID,
		&updatedMilestone.Title,
		&updatedMilestone.Description,
		&updatedMilestone.DueDate,
		&updatedMilestone.CompletedDate,
		&updatedMilestone.IsCompleted,
		&updatedMilestone.Status,
		&updatedMilestone.ProgressPercentage,
		&updatedMilestone.CreatedAt,
		&updatedMilestone.UpdatedAt,
		&updatedMilestone.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("milestone with ID %s not found", milestone.ID)
		}
		return nil, fmt.Errorf("failed to update milestone: %w", err)
	}

	return &updatedMilestone, nil
}

func (r *DB) DeleteMilestone(milestoneID uuid.UUID) error {
	// Use DELETE with RETURNING to get milestone info in one query
	query := `DELETE FROM milestones WHERE id = $1 RETURNING title`
	
	var milestoneTitle string
	err := r.db.QueryRow(query, milestoneID).Scan(&milestoneTitle)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("milestone with ID %s not found", milestoneID)
		}
		return fmt.Errorf("failed to delete milestone: %w", err)
	}

	return nil
}

func (r *DB) GetMilestoneByID(milestoneID uuid.UUID) (*domain.Milestone, error) {
	query := `
		SELECT id, project_id, title, description, due_date, completed_date,
			   isCompleted, status, progress_percentage, created_at, updated_at, created_by
		FROM milestones
		WHERE id = $1`

	var milestone domain.Milestone
	err := r.db.QueryRow(query, milestoneID).Scan(
		&milestone.ID,
		&milestone.ProjectID,
		&milestone.Title,
		&milestone.Description,
		&milestone.DueDate,
		&milestone.CompletedDate,
		&milestone.IsCompleted,
		&milestone.Status,
		&milestone.ProgressPercentage,
		&milestone.CreatedAt,
		&milestone.UpdatedAt,
		&milestone.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("milestone with ID %s not found", milestoneID)
		}
		return nil, fmt.Errorf("failed to get milestone: %w", err)
	}

	return &milestone, nil
}

func (r *DB) GetProjectMilestones(projectID uuid.UUID, filters map[string]interface{}) ([]*domain.Milestone, error) {
	// Build the WHERE clause based on filters
	whereConditions := []string{"project_id = $1"}
	args := []interface{}{projectID}
	argIndex := 2

	for key, value := range filters {
		switch key {
		case "status":
			whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "isCompleted":
			whereConditions = append(whereConditions, fmt.Sprintf("isCompleted = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "overdue":
			if value.(bool) {
				whereConditions = append(whereConditions, "due_date < CURRENT_DATE AND isCompleted = false")
			}
		case "upcoming_days":
			days := value.(int)
			whereConditions = append(whereConditions, fmt.Sprintf("due_date BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '%d days' AND isCompleted = false", days))
		}
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := fmt.Sprintf(`
		SELECT id, project_id, title, description, due_date, completed_date,
			   isCompleted, status, progress_percentage, created_at, updated_at, created_by
		FROM milestones
		WHERE %s
		ORDER BY due_date ASC`, whereClause)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query project milestones: %w", err)
	}
	defer rows.Close()

	var milestones []*domain.Milestone
	for rows.Next() {
		var milestone domain.Milestone
		err := rows.Scan(
			&milestone.ID,
			&milestone.ProjectID,
			&milestone.Title,
			&milestone.Description,
			&milestone.DueDate,
			&milestone.CompletedDate,
			&milestone.IsCompleted,
			&milestone.Status,
			&milestone.ProgressPercentage,
			&milestone.CreatedAt,
			&milestone.UpdatedAt,
			&milestone.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan milestone: %w", err)
		}
		milestones = append(milestones, &milestone)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return milestones, nil
}

func (r *DB) GetAllMilestones(limit, offset int, filters map[string]interface{}) ([]*domain.Milestone, int, error) {
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
		case "status":
			whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "isCompleted":
			whereConditions = append(whereConditions, fmt.Sprintf("isCompleted = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "overdue":
			if value.(bool) {
				whereConditions = append(whereConditions, "due_date < CURRENT_DATE AND isCompleted = false")
			}
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM milestones %s", whereClause)
	var totalCount int
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count milestones: %w", err)
	}

	// Get milestones with pagination
	query := fmt.Sprintf(`
		SELECT id, project_id, title, description, due_date, completed_date,
			   isCompleted, status, progress_percentage, created_at, updated_at, created_by
		FROM milestones
		%s
		ORDER BY due_date ASC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query milestones: %w", err)
	}
	defer rows.Close()

	var milestones []*domain.Milestone
	for rows.Next() {
		var milestone domain.Milestone
		err := rows.Scan(
			&milestone.ID,
			&milestone.ProjectID,
			&milestone.Title,
			&milestone.Description,
			&milestone.DueDate,
			&milestone.CompletedDate,
			&milestone.IsCompleted,
			&milestone.Status,
			&milestone.ProgressPercentage,
			&milestone.CreatedAt,
			&milestone.UpdatedAt,
			&milestone.CreatedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan milestone: %w", err)
		}
		milestones = append(milestones, &milestone)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return milestones, totalCount, nil
}

func (r *DB) CompleteMilestone(milestoneID uuid.UUID) (*domain.Milestone, error) {
	query := `
		UPDATE milestones SET
			isCompleted = true,
			completed_date = CURRENT_DATE,
			status = 'completed',
			progress_percentage = 100,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, project_id, title, description, due_date, completed_date,
				 isCompleted, status, progress_percentage, created_at, updated_at, created_by`

	var milestone domain.Milestone
	err := r.db.QueryRow(query, milestoneID).Scan(
		&milestone.ID,
		&milestone.ProjectID,
		&milestone.Title,
		&milestone.Description,
		&milestone.DueDate,
		&milestone.CompletedDate,
		&milestone.IsCompleted,
		&milestone.Status,
		&milestone.ProgressPercentage,
		&milestone.CreatedAt,
		&milestone.UpdatedAt,
		&milestone.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("milestone with ID %s not found", milestoneID)
		}
		return nil, fmt.Errorf("failed to complete milestone: %w", err)
	}

	return &milestone, nil
}

func (r *DB) GetOverdueMilestones(projectID *uuid.UUID) ([]*domain.Milestone, error) {
	query := `
		SELECT id, project_id, title, description, due_date, completed_date,
			   isCompleted, status, progress_percentage, created_at, updated_at, created_by
		FROM milestones
		WHERE due_date < CURRENT_DATE AND isCompleted = false`
	
	args := []interface{}{}
	
	if projectID != nil {
		query += " AND project_id = $1"
		args = append(args, *projectID)
	}
	
	query += " ORDER BY due_date ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query overdue milestones: %w", err)
	}
	defer rows.Close()

	var milestones []*domain.Milestone
	for rows.Next() {
		var milestone domain.Milestone
		err := rows.Scan(
			&milestone.ID,
			&milestone.ProjectID,
			&milestone.Title,
			&milestone.Description,
			&milestone.DueDate,
			&milestone.CompletedDate,
			&milestone.IsCompleted,
			&milestone.Status,
			&milestone.ProgressPercentage,
			&milestone.CreatedAt,
			&milestone.UpdatedAt,
			&milestone.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan overdue milestone: %w", err)
		}
		milestones = append(milestones, &milestone)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return milestones, nil
}

func (r *DB) GetUpcomingMilestones(projectID uuid.UUID, days int) ([]*domain.Milestone, error) {
	query := `
		SELECT id, project_id, title, description, due_date, completed_date,
			   isCompleted, status, progress_percentage, created_at, updated_at, created_by
		FROM milestones
		WHERE project_id = $1 
		  AND due_date BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '%d days'
		  AND isCompleted = false
		ORDER BY due_date ASC`

	formattedQuery := fmt.Sprintf(query, days)

	rows, err := r.db.Query(formattedQuery, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query upcoming milestones: %w", err)
	}
	defer rows.Close()

	var milestones []*domain.Milestone
	for rows.Next() {
		var milestone domain.Milestone
		err := rows.Scan(
			&milestone.ID,
			&milestone.ProjectID,
			&milestone.Title,
			&milestone.Description,
			&milestone.DueDate,
			&milestone.CompletedDate,
			&milestone.IsCompleted,
			&milestone.Status,
			&milestone.ProgressPercentage,
			&milestone.CreatedAt,
			&milestone.UpdatedAt,
			&milestone.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan upcoming milestone: %w", err)
		}
		milestones = append(milestones, &milestone)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return milestones, nil
}

func (r *DB) UpdateMilestoneProgress(milestoneID uuid.UUID, progressPercentage int) (*domain.Milestone, error) {
	// Determine status based on progress
	var status string
	switch {
	case progressPercentage == 100:
		status = "completed"
	case progressPercentage > 0:
		status = "in_progress"
	default:
		status = "pending"
	}

	query := `
		UPDATE milestones SET
			progress_percentage = $2,
			status = $3,
			isCompleted = $4,
			completed_date = CASE WHEN $4 = true THEN CURRENT_DATE ELSE completed_date END,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, project_id, title, description, due_date, completed_date,
				 isCompleted, status, progress_percentage, created_at, updated_at, created_by`

	isCompleted := progressPercentage == 100

	var milestone domain.Milestone
	err := r.db.QueryRow(query, milestoneID, progressPercentage, status, isCompleted).Scan(
		&milestone.ID,
		&milestone.ProjectID,
		&milestone.Title,
		&milestone.Description,
		&milestone.DueDate,
		&milestone.CompletedDate,
		&milestone.IsCompleted,
		&milestone.Status,
		&milestone.ProgressPercentage,
		&milestone.CreatedAt,
		&milestone.UpdatedAt,
		&milestone.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("milestone with ID %s not found", milestoneID)
		}
		return nil, fmt.Errorf("failed to update milestone progress: %w", err)
	}

	return &milestone, nil
}
