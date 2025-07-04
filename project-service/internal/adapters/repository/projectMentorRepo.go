package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
)

func (r *DB) AssignMentor(projectMentor *domain.ProjectMentor) (*domain.ProjectMentor, error) {
	query := `
		INSERT INTO project_mentors (
			id, project_id, mentor_id, mentorship_type, status, start_date,
			end_date, hours_committed, is_active, assigned_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
		RETURNING id, project_id, mentor_id, mentorship_type, status, start_date,
				 end_date, hours_committed, is_active, assigned_by, created_at, updated_at`

	// Generate UUID if not provided
	if projectMentor.ID == uuid.Nil {
		projectMentor.ID = uuid.New()
	}

	var createdProjectMentor domain.ProjectMentor
	err := r.db.QueryRow(
		query,
		projectMentor.ID,
		projectMentor.ProjectID,
		projectMentor.MentorID,
		projectMentor.MentorshipType,
		projectMentor.Status,
		projectMentor.StartDate,
		projectMentor.EndDate,
		projectMentor.HoursCommitted,
		projectMentor.IsActive,
		projectMentor.AssignedBy,
	).Scan(
		&createdProjectMentor.ID,
		&createdProjectMentor.ProjectID,
		&createdProjectMentor.MentorID,
		&createdProjectMentor.MentorshipType,
		&createdProjectMentor.Status,
		&createdProjectMentor.StartDate,
		&createdProjectMentor.EndDate,
		&createdProjectMentor.HoursCommitted,
		&createdProjectMentor.IsActive,
		&createdProjectMentor.AssignedBy,
		&createdProjectMentor.CreatedAt,
		&createdProjectMentor.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to assign mentor: %w", err)
	}

	return &createdProjectMentor, nil
}

func (r *DB) UpdateProjectMentor(projectMentor *domain.ProjectMentor) (*domain.ProjectMentor, error) {
	query := `
		UPDATE project_mentors SET
			mentorship_type = $2,
			status = $3,
			start_date = $4,
			end_date = $5,
			hours_committed = $6,
			is_active = $7,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, project_id, mentor_id, mentorship_type, status, start_date,
				 end_date, hours_committed, is_active, assigned_by, created_at, updated_at`

	var updatedProjectMentor domain.ProjectMentor
	err := r.db.QueryRow(
		query,
		projectMentor.ID,
		projectMentor.MentorshipType,
		projectMentor.Status,
		projectMentor.StartDate,
		projectMentor.EndDate,
		projectMentor.HoursCommitted,
		projectMentor.IsActive,
	).Scan(
		&updatedProjectMentor.ID,
		&updatedProjectMentor.ProjectID,
		&updatedProjectMentor.MentorID,
		&updatedProjectMentor.MentorshipType,
		&updatedProjectMentor.Status,
		&updatedProjectMentor.StartDate,
		&updatedProjectMentor.EndDate,
		&updatedProjectMentor.HoursCommitted,
		&updatedProjectMentor.IsActive,
		&updatedProjectMentor.AssignedBy,
		&updatedProjectMentor.CreatedAt,
		&updatedProjectMentor.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project mentor with ID %s not found", projectMentor.ID)
		}
		return nil, fmt.Errorf("failed to update project mentor: %w", err)
	}

	return &updatedProjectMentor, nil
}

func (r *DB) RemoveMentor(projectMentorID uuid.UUID) error {
	// Use DELETE with RETURNING to get project mentor info in one query
	query := `DELETE FROM project_mentors WHERE id = $1 RETURNING id`
	
	var deletedID uuid.UUID
	err := r.db.QueryRow(query, projectMentorID).Scan(&deletedID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("project mentor with ID %s not found", projectMentorID)
		}
		return fmt.Errorf("failed to remove mentor: %w", err)
	}

	return nil
}

func (r *DB) GetProjectMentorByID(projectMentorID uuid.UUID) (*domain.ProjectMentor, error) {
	query := `
		SELECT id, project_id, mentor_id, mentorship_type, status, start_date,
			   end_date, hours_committed, is_active, assigned_by, created_at, updated_at
		FROM project_mentors
		WHERE id = $1`

	var projectMentor domain.ProjectMentor
	err := r.db.QueryRow(query, projectMentorID).Scan(
		&projectMentor.ID,
		&projectMentor.ProjectID,
		&projectMentor.MentorID,
		&projectMentor.MentorshipType,
		&projectMentor.Status,
		&projectMentor.StartDate,
		&projectMentor.EndDate,
		&projectMentor.HoursCommitted,
		&projectMentor.IsActive,
		&projectMentor.AssignedBy,
		&projectMentor.CreatedAt,
		&projectMentor.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project mentor with ID %s not found", projectMentorID)
		}
		return nil, fmt.Errorf("failed to get project mentor: %w", err)
	}

	return &projectMentor, nil
}

func (r *DB) GetProjectMentors(projectID uuid.UUID, activeOnly bool) ([]*domain.ProjectMentor, error) {
	query := `
		SELECT id, project_id, mentor_id, mentorship_type, status, start_date,
			   end_date, hours_committed, is_active, assigned_by, created_at, updated_at
		FROM project_mentors
		WHERE project_id = $1`
	
	args := []interface{}{projectID}
	
	if activeOnly {
		query += " AND is_active = true"
	}
	
	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query project mentors: %w", err)
	}
	defer rows.Close()

	var projectMentors []*domain.ProjectMentor
	for rows.Next() {
		var projectMentor domain.ProjectMentor
		err := rows.Scan(
			&projectMentor.ID,
			&projectMentor.ProjectID,
			&projectMentor.MentorID,
			&projectMentor.MentorshipType,
			&projectMentor.Status,
			&projectMentor.StartDate,
			&projectMentor.EndDate,
			&projectMentor.HoursCommitted,
			&projectMentor.IsActive,
			&projectMentor.AssignedBy,
			&projectMentor.CreatedAt,
			&projectMentor.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project mentor: %w", err)
		}
		projectMentors = append(projectMentors, &projectMentor)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return projectMentors, nil
}

func (r *DB) GetMentorProjects(mentorID uuid.UUID, activeOnly bool) ([]*domain.ProjectMentor, error) {
	query := `
		SELECT id, project_id, mentor_id, mentorship_type, status, start_date,
			   end_date, hours_committed, is_active, assigned_by, created_at, updated_at
		FROM project_mentors
		WHERE mentor_id = $1`
	
	args := []interface{}{mentorID}
	
	if activeOnly {
		query += " AND is_active = true"
	}
	
	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query mentor projects: %w", err)
	}
	defer rows.Close()

	var projectMentors []*domain.ProjectMentor
	for rows.Next() {
		var projectMentor domain.ProjectMentor
		err := rows.Scan(
			&projectMentor.ID,
			&projectMentor.ProjectID,
			&projectMentor.MentorID,
			&projectMentor.MentorshipType,
			&projectMentor.Status,
			&projectMentor.StartDate,
			&projectMentor.EndDate,
			&projectMentor.HoursCommitted,
			&projectMentor.IsActive,
			&projectMentor.AssignedBy,
			&projectMentor.CreatedAt,
			&projectMentor.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan mentor project: %w", err)
		}
		projectMentors = append(projectMentors, &projectMentor)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return projectMentors, nil
}

func (r *DB) GetAllProjectMentors(limit, offset int, filters map[string]interface{}) ([]*domain.ProjectMentor, int, error) {
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
		case "mentor_id":
			whereConditions = append(whereConditions, fmt.Sprintf("mentor_id = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "mentorship_type":
			whereConditions = append(whereConditions, fmt.Sprintf("mentorship_type = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "status":
			whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "is_active":
			whereConditions = append(whereConditions, fmt.Sprintf("is_active = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "assigned_by":
			whereConditions = append(whereConditions, fmt.Sprintf("assigned_by = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "recent_days":
			days := value.(int)
			whereConditions = append(whereConditions, fmt.Sprintf("created_at >= CURRENT_DATE - INTERVAL '%d days'", days))
		}
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM project_mentors %s", whereClause)
	var totalCount int
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count project mentors: %w", err)
	}

	// Get project mentors with pagination
	query := fmt.Sprintf(`
		SELECT id, project_id, mentor_id, mentorship_type, status, start_date,
			   end_date, hours_committed, is_active, assigned_by, created_at, updated_at
		FROM project_mentors
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query project mentors: %w", err)
	}
	defer rows.Close()

	var projectMentors []*domain.ProjectMentor
	for rows.Next() {
		var projectMentor domain.ProjectMentor
		err := rows.Scan(
			&projectMentor.ID,
			&projectMentor.ProjectID,
			&projectMentor.MentorID,
			&projectMentor.MentorshipType,
			&projectMentor.Status,
			&projectMentor.StartDate,
			&projectMentor.EndDate,
			&projectMentor.HoursCommitted,
			&projectMentor.IsActive,
			&projectMentor.AssignedBy,
			&projectMentor.CreatedAt,
			&projectMentor.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan project mentor: %w", err)
		}
		projectMentors = append(projectMentors, &projectMentor)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return projectMentors, totalCount, nil
}

func (r *DB) CompleteMentorship(projectMentorID uuid.UUID) (*domain.ProjectMentor, error) {
	query := `
		UPDATE project_mentors SET
			status = 'completed',
			is_active = false,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, project_id, mentor_id, mentorship_type, status, start_date,
				 end_date, hours_committed, is_active, assigned_by, created_at, updated_at`

	var projectMentor domain.ProjectMentor
	err := r.db.QueryRow(query, projectMentorID).Scan(
		&projectMentor.ID,
		&projectMentor.ProjectID,
		&projectMentor.MentorID,
		&projectMentor.MentorshipType,
		&projectMentor.Status,
		&projectMentor.StartDate,
		&projectMentor.EndDate,
		&projectMentor.HoursCommitted,
		&projectMentor.IsActive,
		&projectMentor.AssignedBy,
		&projectMentor.CreatedAt,
		&projectMentor.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project mentor with ID %s not found", projectMentorID)
		}
		return nil, fmt.Errorf("failed to complete mentorship: %w", err)
	}

	return &projectMentor, nil
}

func (r *DB) UpdateMentorshipStatus(projectMentorID uuid.UUID, status string) (*domain.ProjectMentor, error) {
	query := `
		UPDATE project_mentors SET
			status = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, project_id, mentor_id, mentorship_type, status, start_date,
				 end_date, hours_committed, is_active, assigned_by, created_at, updated_at`

	var projectMentor domain.ProjectMentor
	err := r.db.QueryRow(query, projectMentorID, status).Scan(
		&projectMentor.ID,
		&projectMentor.ProjectID,
		&projectMentor.MentorID,
		&projectMentor.MentorshipType,
		&projectMentor.Status,
		&projectMentor.StartDate,
		&projectMentor.EndDate,
		&projectMentor.HoursCommitted,
		&projectMentor.IsActive,
		&projectMentor.AssignedBy,
		&projectMentor.CreatedAt,
		&projectMentor.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project mentor with ID %s not found", projectMentorID)
		}
		return nil, fmt.Errorf("failed to update mentorship status: %w", err)
	}

	return &projectMentor, nil
}

func (r *DB) GetMentorshipsByType(mentorshipType string, projectID *uuid.UUID) ([]*domain.ProjectMentor, error) {
	query := `
		SELECT id, project_id, mentor_id, mentorship_type, status, start_date,
			   end_date, hours_committed, is_active, assigned_by, created_at, updated_at
		FROM project_mentors
		WHERE mentorship_type = $1`
	
	args := []interface{}{mentorshipType}
	
	if projectID != nil {
		query += " AND project_id = $2"
		args = append(args, *projectID)
	}
	
	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query mentorships by type: %w", err)
	}
	defer rows.Close()

	var projectMentors []*domain.ProjectMentor
	for rows.Next() {
		var projectMentor domain.ProjectMentor
		err := rows.Scan(
			&projectMentor.ID,
			&projectMentor.ProjectID,
			&projectMentor.MentorID,
			&projectMentor.MentorshipType,
			&projectMentor.Status,
			&projectMentor.StartDate,
			&projectMentor.EndDate,
			&projectMentor.HoursCommitted,
			&projectMentor.IsActive,
			&projectMentor.AssignedBy,
			&projectMentor.CreatedAt,
			&projectMentor.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project mentor: %w", err)
		}
		projectMentors = append(projectMentors, &projectMentor)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return projectMentors, nil
}

func (r *DB) GetMentorshipsByStatus(status string, projectID *uuid.UUID) ([]*domain.ProjectMentor, error) {
	query := `
		SELECT id, project_id, mentor_id, mentorship_type, status, start_date,
			   end_date, hours_committed, is_active, assigned_by, created_at, updated_at
		FROM project_mentors
		WHERE status = $1`
	
	args := []interface{}{status}
	
	if projectID != nil {
		query += " AND project_id = $2"
		args = append(args, *projectID)
	}
	
	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query mentorships by status: %w", err)
	}
	defer rows.Close()

	var projectMentors []*domain.ProjectMentor
	for rows.Next() {
		var projectMentor domain.ProjectMentor
		err := rows.Scan(
			&projectMentor.ID,
			&projectMentor.ProjectID,
			&projectMentor.MentorID,
			&projectMentor.MentorshipType,
			&projectMentor.Status,
			&projectMentor.StartDate,
			&projectMentor.EndDate,
			&projectMentor.HoursCommitted,
			&projectMentor.IsActive,
			&projectMentor.AssignedBy,
			&projectMentor.CreatedAt,
			&projectMentor.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project mentor: %w", err)
		}
		projectMentors = append(projectMentors, &projectMentor)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return projectMentors, nil
}

func (r *DB) GetMentorshipStatistics(projectID *uuid.UUID) (*ports.MentorshipStatistics, error) {
	query := `
		SELECT 
			COUNT(*) as total_mentorships,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_mentorships,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_mentorships,
			COALESCE(SUM(hours_committed), 0) as total_hours_committed,
			COUNT(CASE WHEN created_at >= CURRENT_DATE - INTERVAL '30 days' THEN 1 END) as recent_mentorships
		FROM project_mentors`
	
	args := []interface{}{}
	
	if projectID != nil {
		query += " WHERE project_id = $1"
		args = append(args, *projectID)
	}

	var stats ports.MentorshipStatistics
	var totalHoursCommitted sql.NullInt64

	err := r.db.QueryRow(query, args...).Scan(
		&stats.TotalMentorships,
		&stats.ActiveMentorships,
		&stats.CompletedMentorships,
		&totalHoursCommitted,
		&stats.RecentMentorships,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get mentorship statistics: %w", err)
	}

	stats.TotalHoursCommitted = int(totalHoursCommitted.Int64)
	if stats.TotalMentorships > 0 {
		stats.AverageHoursCommitted = float64(stats.TotalHoursCommitted) / float64(stats.TotalMentorships)
	}

	// Get mentorships by type
	typeQuery := `
		SELECT mentorship_type, COUNT(*) 
		FROM project_mentors`
	
	if projectID != nil {
		typeQuery += " WHERE project_id = $1"
	}
	
	typeQuery += " GROUP BY mentorship_type"

	typeRows, err := r.db.Query(typeQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get mentorship type statistics: %w", err)
	}
	defer typeRows.Close()

	stats.MentorshipsByType = make(map[string]int)
	for typeRows.Next() {
		var mentorshipType string
		var count int
		err := typeRows.Scan(&mentorshipType, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan mentorship type: %w", err)
		}
		stats.MentorshipsByType[mentorshipType] = count
	}

	// Get mentorships by status
	statusQuery := `
		SELECT status, COUNT(*) 
		FROM project_mentors`
	
	if projectID != nil {
		statusQuery += " WHERE project_id = $1"
	}
	
	statusQuery += " GROUP BY status"

	statusRows, err := r.db.Query(statusQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get mentorship status statistics: %w", err)
	}
	defer statusRows.Close()

	stats.MentorshipsByStatus = make(map[string]int)
	for statusRows.Next() {
		var status string
		var count int
		err := statusRows.Scan(&status, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan mentorship status: %w", err)
		}
		stats.MentorshipsByStatus[status] = count
	}

	// Get mentorships by assigner
	assignerQuery := `
		SELECT assigned_by::text, COUNT(*) 
		FROM project_mentors`
	
	if projectID != nil {
		assignerQuery += " WHERE project_id = $1"
	}
	
	assignerQuery += " GROUP BY assigned_by"

	assignerRows, err := r.db.Query(assignerQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get assigner statistics: %w", err)
	}
	defer assignerRows.Close()

	stats.MentorshipsByAssigner = make(map[string]int)
	for assignerRows.Next() {
		var assignerID string
		var count int
		err := assignerRows.Scan(&assignerID, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assigner: %w", err)
		}
		stats.MentorshipsByAssigner[assignerID] = count
	}

	return &stats, nil
}

func (r *DB) IsProjectMentorshipExists(projectID, mentorID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM project_mentors 
			WHERE project_id = $1 AND mentor_id = $2
		)`

	var exists bool
	err := r.db.QueryRow(query, projectID, mentorID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check project mentorship existence: %w", err)
	}

	return exists, nil
}
