package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
)

func (r *DB) AddProjectMember(member *domain.ProjectMember) (*domain.ProjectMember, error) {
	query := `
		INSERT INTO project_members (
			id, project_id, user_id, role, is_active, joined_date,
			can_edit_project, can_manage_tasks, can_view_reports
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
		RETURNING id, project_id, user_id, role, is_active, joined_date,
				 left_date, can_edit_project, can_manage_tasks, can_view_reports,
				 created_at, updated_at`

	// Generate UUID if not provided
	if member.ID == uuid.Nil {
		member.ID = uuid.New()
	}

	// Set joined_date to current date if not provided
	if member.JoinedDate.IsZero() {
		member.JoinedDate = time.Now()
	}

	var createdMember domain.ProjectMember
	err := r.db.QueryRow(
		query,
		member.ID,
		member.ProjectID,
		member.UserID,
		member.Role,
		member.IsActive,
		member.JoinedDate,
		member.CanEditProject,
		member.CanManageTasks,
		member.CanViewReports,
	).Scan(
		&createdMember.ID,
		&createdMember.ProjectID,
		&createdMember.UserID,
		&createdMember.Role,
		&createdMember.IsActive,
		&createdMember.JoinedDate,
		&createdMember.LeftDate,
		&createdMember.CanEditProject,
		&createdMember.CanManageTasks,
		&createdMember.CanViewReports,
		&createdMember.CreatedAt,
		&createdMember.UpdatedAt,
	)

	if err != nil {
		// Check for unique constraint violation
		if err.Error() == `pq: duplicate key value violates unique constraint "unique_project_user"` {
			return nil, fmt.Errorf("user is already a member of this project")
		}
		return nil, fmt.Errorf("failed to add project member: %w", err)
	}

	return &createdMember, nil
}

func (r *DB) UpdateProjectMember(member *domain.ProjectMember) (*domain.ProjectMember, error) {
	query := `
		UPDATE project_members SET
			role = $2,
			is_active = $3,
			left_date = $4,
			can_edit_project = $5,
			can_manage_tasks = $6,
			can_view_reports = $7,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, project_id, user_id, role, is_active, joined_date,
				 left_date, can_edit_project, can_manage_tasks, can_view_reports,
				 created_at, updated_at`

	var updatedMember domain.ProjectMember
	err := r.db.QueryRow(
		query,
		member.ID,
		member.Role,
		member.IsActive,
		member.LeftDate,
		member.CanEditProject,
		member.CanManageTasks,
		member.CanViewReports,
	).Scan(
		&updatedMember.ID,
		&updatedMember.ProjectID,
		&updatedMember.UserID,
		&updatedMember.Role,
		&updatedMember.IsActive,
		&updatedMember.JoinedDate,
		&updatedMember.LeftDate,
		&updatedMember.CanEditProject,
		&updatedMember.CanManageTasks,
		&updatedMember.CanViewReports,
		&updatedMember.CreatedAt,
		&updatedMember.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project member with ID %s not found", member.ID)
		}
		return nil, fmt.Errorf("failed to update project member: %w", err)
	}

	return &updatedMember, nil
}

func (r *DB) RemoveProjectMember(memberID uuid.UUID) error {
	// Use DELETE with RETURNING to get member info in one query
	query := `DELETE FROM project_members WHERE id = $1 RETURNING role`
	
	var memberRole string
	err := r.db.QueryRow(query, memberID).Scan(&memberRole)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("project member with ID %s not found", memberID)
		}
		return fmt.Errorf("failed to remove project member: %w", err)
	}

	return nil
}

func (r *DB) GetProjectMemberByID(memberID uuid.UUID) (*domain.ProjectMember, error) {
	query := `
		SELECT id, project_id, user_id, role, is_active, joined_date,
			   left_date, can_edit_project, can_manage_tasks, can_view_reports,
			   created_at, updated_at
		FROM project_members
		WHERE id = $1`

	var member domain.ProjectMember
	err := r.db.QueryRow(query, memberID).Scan(
		&member.ID,
		&member.ProjectID,
		&member.UserID,
		&member.Role,
		&member.IsActive,
		&member.JoinedDate,
		&member.LeftDate,
		&member.CanEditProject,
		&member.CanManageTasks,
		&member.CanViewReports,
		&member.CreatedAt,
		&member.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project member with ID %s not found", memberID)
		}
		return nil, fmt.Errorf("failed to get project member: %w", err)
	}

	return &member, nil
}

func (r *DB) GetProjectMembers(projectID uuid.UUID, activeOnly bool) ([]*domain.ProjectMember, error) {
	query := `
		SELECT id, project_id, user_id, role, is_active, joined_date,
			   left_date, can_edit_project, can_manage_tasks, can_view_reports,
			   created_at, updated_at
		FROM project_members
		WHERE project_id = $1`
	
	args := []interface{}{projectID}
	
	if activeOnly {
		query += " AND is_active = true"
	}
	
	query += " ORDER BY joined_date ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query project members: %w", err)
	}
	defer rows.Close()

	var members []*domain.ProjectMember
	for rows.Next() {
		var member domain.ProjectMember
		err := rows.Scan(
			&member.ID,
			&member.ProjectID,
			&member.UserID,
			&member.Role,
			&member.IsActive,
			&member.JoinedDate,
			&member.LeftDate,
			&member.CanEditProject,
			&member.CanManageTasks,
			&member.CanViewReports,
			&member.CreatedAt,
			&member.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project member: %w", err)
		}
		members = append(members, &member)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return members, nil
}

func (r *DB) GetUserProjects(userID uuid.UUID, activeOnly bool) ([]*domain.ProjectMember, error) {
	query := `
		SELECT id, project_id, user_id, role, is_active, joined_date,
			   left_date, can_edit_project, can_manage_tasks, can_view_reports,
			   created_at, updated_at
		FROM project_members
		WHERE user_id = $1`
	
	args := []interface{}{userID}
	
	if activeOnly {
		query += " AND is_active = true"
	}
	
	query += " ORDER BY joined_date DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query user projects: %w", err)
	}
	defer rows.Close()

	var members []*domain.ProjectMember
	for rows.Next() {
		var member domain.ProjectMember
		err := rows.Scan(
			&member.ID,
			&member.ProjectID,
			&member.UserID,
			&member.Role,
			&member.IsActive,
			&member.JoinedDate,
			&member.LeftDate,
			&member.CanEditProject,
			&member.CanManageTasks,
			&member.CanViewReports,
			&member.CreatedAt,
			&member.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user project: %w", err)
		}
		members = append(members, &member)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return members, nil
}

func (r *DB) IsUserMemberOfProject(projectID, userID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM project_members 
			WHERE project_id = $1 AND user_id = $2 AND is_active = true
		)`

	var exists bool
	err := r.db.QueryRow(query, projectID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check project membership: %w", err)
	}

	return exists, nil
}

func (r *DB) GetUserProjectMembership(userID, projectID uuid.UUID) (*domain.ProjectMember, error) {
	query := `
		SELECT id, project_id, user_id, role, is_active, joined_date,
			   left_date, can_edit_project, can_manage_tasks, can_view_reports,
			   created_at, updated_at
		FROM project_members
		WHERE user_id = $1 AND project_id = $2`

	var member domain.ProjectMember
	err := r.db.QueryRow(query, userID, projectID).Scan(
		&member.ID,
		&member.ProjectID,
		&member.UserID,
		&member.Role,
		&member.IsActive,
		&member.JoinedDate,
		&member.LeftDate,
		&member.CanEditProject,
		&member.CanManageTasks,
		&member.CanViewReports,
		&member.CreatedAt,
		&member.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User is not a member of this project
		}
		return nil, fmt.Errorf("failed to get user project membership: %w", err)
	}

	return &member, nil
}
