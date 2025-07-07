package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
)

func (r *DB) CreateMentor(mentor *domain.Mentor) (*domain.Mentor, error) {
	query := `
		INSERT INTO mentors (
			id, user_id, company, position, expertise_area, years_experience,
			availability_type, linkedin_url, website_url, phone, is_verified, is_active
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
		RETURNING id, user_id, company, position, expertise_area, years_experience,
				 availability_type, linkedin_url, website_url, phone, is_verified, is_active,
				 created_at, updated_at`

	// Generate UUID if not provided
	if mentor.ID == uuid.Nil {
		mentor.ID = uuid.New()
	}

	var createdMentor domain.Mentor
	err := r.db.QueryRow(
		query,
		mentor.ID,
		mentor.UserID,
		mentor.Company,
		mentor.Position,
		mentor.ExpertiseArea,
		mentor.YearsExperience,
		mentor.AvailabilityType,
		mentor.LinkedinURL,
		mentor.WebsiteURL,
		mentor.Phone,
		mentor.IsVerified,
		mentor.IsActive,
	).Scan(
		&createdMentor.ID,
		&createdMentor.UserID,
		&createdMentor.Company,
		&createdMentor.Position,
		&createdMentor.ExpertiseArea,
		&createdMentor.YearsExperience,
		&createdMentor.AvailabilityType,
		&createdMentor.LinkedinURL,
		&createdMentor.WebsiteURL,
		&createdMentor.Phone,
		&createdMentor.IsVerified,
		&createdMentor.IsActive,
		&createdMentor.CreatedAt,
		&createdMentor.UpdatedAt,
	)

	if err != nil {
		// Check for unique constraint violation on user_id
		if strings.Contains(err.Error(), "unique constraint") && strings.Contains(err.Error(), "user_id") {
			return nil, fmt.Errorf("user is already registered as a mentor")
		}
		return nil, fmt.Errorf("failed to create mentor: %w", err)
	}

	return &createdMentor, nil
}

func (r *DB) UpdateMentor(mentor *domain.Mentor) (*domain.Mentor, error) {
	query := `
		UPDATE mentors SET
			company = $2,
			position = $3,
			expertise_area = $4,
			years_experience = $5,
			availability_type = $6,
			linkedin_url = $7,
			website_url = $8,
			phone = $9,
			is_verified = $10,
			is_active = $11,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, user_id, company, position, expertise_area, years_experience,
				 availability_type, linkedin_url, website_url, phone, is_verified, is_active,
				 created_at, updated_at`

	var updatedMentor domain.Mentor
	err := r.db.QueryRow(
		query,
		mentor.ID,
		mentor.Company,
		mentor.Position,
		mentor.ExpertiseArea,
		mentor.YearsExperience,
		mentor.AvailabilityType,
		mentor.LinkedinURL,
		mentor.WebsiteURL,
		mentor.Phone,
		mentor.IsVerified,
		mentor.IsActive,
	).Scan(
		&updatedMentor.ID,
		&updatedMentor.UserID,
		&updatedMentor.Company,
		&updatedMentor.Position,
		&updatedMentor.ExpertiseArea,
		&updatedMentor.YearsExperience,
		&updatedMentor.AvailabilityType,
		&updatedMentor.LinkedinURL,
		&updatedMentor.WebsiteURL,
		&updatedMentor.Phone,
		&updatedMentor.IsVerified,
		&updatedMentor.IsActive,
		&updatedMentor.CreatedAt,
		&updatedMentor.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("mentor with ID %s not found", mentor.ID)
		}
		return nil, fmt.Errorf("failed to update mentor: %w", err)
	}

	return &updatedMentor, nil
}

func (r *DB) DeleteMentor(mentorID uuid.UUID) error {
	// Use DELETE with RETURNING to get mentor info in one query
	query := `DELETE FROM mentors WHERE id = $1 RETURNING user_id`
	
	var userID uuid.UUID
	err := r.db.QueryRow(query, mentorID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("mentor with ID %s not found", mentorID)
		}
		return fmt.Errorf("failed to delete mentor: %w", err)
	}

	return nil
}

func (r *DB) GetMentorByID(mentorID uuid.UUID) (*domain.Mentor, error) {
	query := `
		SELECT id, user_id, company, position, expertise_area, years_experience,
			   availability_type, linkedin_url, website_url, phone, is_verified, is_active,
			   created_at, updated_at
		FROM mentors
		WHERE id = $1`

	var mentor domain.Mentor
	err := r.db.QueryRow(query, mentorID).Scan(
		&mentor.ID,
		&mentor.UserID,
		&mentor.Company,
		&mentor.Position,
		&mentor.ExpertiseArea,
		&mentor.YearsExperience,
		&mentor.AvailabilityType,
		&mentor.LinkedinURL,
		&mentor.WebsiteURL,
		&mentor.Phone,
		&mentor.IsVerified,
		&mentor.IsActive,
		&mentor.CreatedAt,
		&mentor.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("mentor with ID %s not found", mentorID)
		}
		return nil, fmt.Errorf("failed to get mentor: %w", err)
	}

	return &mentor, nil
}

func (r *DB) GetMentorByUserID(userID uuid.UUID) (*domain.Mentor, error) {
	query := `
		SELECT id, user_id, company, position, expertise_area, years_experience,
			   availability_type, linkedin_url, website_url, phone, is_verified, is_active,
			   created_at, updated_at
		FROM mentors
		WHERE user_id = $1`

	var mentor domain.Mentor
	err := r.db.QueryRow(query, userID).Scan(
		&mentor.ID,
		&mentor.UserID,
		&mentor.Company,
		&mentor.Position,
		&mentor.ExpertiseArea,
		&mentor.YearsExperience,
		&mentor.AvailabilityType,
		&mentor.LinkedinURL,
		&mentor.WebsiteURL,
		&mentor.Phone,
		&mentor.IsVerified,
		&mentor.IsActive,
		&mentor.CreatedAt,
		&mentor.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("mentor with user ID %s not found", userID)
		}
		return nil, fmt.Errorf("failed to get mentor by user ID: %w", err)
	}

	return &mentor, nil
}

func (r *DB) GetAllMentors(limit, offset int, filters map[string]interface{}) ([]*domain.Mentor, int, error) {
	// Build the WHERE clause based on filters
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	for key, value := range filters {
		switch key {
		case "expertise_area":
			whereConditions = append(whereConditions, fmt.Sprintf("expertise_area = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "availability_type":
			whereConditions = append(whereConditions, fmt.Sprintf("availability_type = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "is_verified":
			whereConditions = append(whereConditions, fmt.Sprintf("is_verified = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "is_active":
			whereConditions = append(whereConditions, fmt.Sprintf("is_active = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "min_experience":
			whereConditions = append(whereConditions, fmt.Sprintf("years_experience >= $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "company":
			whereConditions = append(whereConditions, fmt.Sprintf("company ILIKE $%d", argIndex))
			args = append(args, fmt.Sprintf("%%%s%%", value))
			argIndex++
		}
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM mentors %s", whereClause)
	var totalCount int
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count mentors: %w", err)
	}

	// Get mentors with pagination
	query := fmt.Sprintf(`
		SELECT id, user_id, company, position, expertise_area, years_experience,
			   availability_type, linkedin_url, website_url, phone, is_verified, is_active,
			   created_at, updated_at
		FROM mentors
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query mentors: %w", err)
	}
	defer rows.Close()

	var mentors []*domain.Mentor
	for rows.Next() {
		var mentor domain.Mentor
		err := rows.Scan(
			&mentor.ID,
			&mentor.UserID,
			&mentor.Company,
			&mentor.Position,
			&mentor.ExpertiseArea,
			&mentor.YearsExperience,
			&mentor.AvailabilityType,
			&mentor.LinkedinURL,
			&mentor.WebsiteURL,
			&mentor.Phone,
			&mentor.IsVerified,
			&mentor.IsActive,
			&mentor.CreatedAt,
			&mentor.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan mentor: %w", err)
		}
		mentors = append(mentors, &mentor)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return mentors, totalCount, nil
}

func (r *DB) DeactivateMentor(mentorID uuid.UUID) (*domain.Mentor, error) {
	query := `
		UPDATE mentors SET
			is_active = false,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, user_id, company, position, expertise_area, years_experience,
				 availability_type, linkedin_url, website_url, phone, is_verified, is_active,
				 created_at, updated_at`

	var mentor domain.Mentor
	err := r.db.QueryRow(query, mentorID).Scan(
		&mentor.ID,
		&mentor.UserID,
		&mentor.Company,
		&mentor.Position,
		&mentor.ExpertiseArea,
		&mentor.YearsExperience,
		&mentor.AvailabilityType,
		&mentor.LinkedinURL,
		&mentor.WebsiteURL,
		&mentor.Phone,
		&mentor.IsVerified,
		&mentor.IsActive,
		&mentor.CreatedAt,
		&mentor.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("mentor with ID %s not found", mentorID)
		}
		return nil, fmt.Errorf("failed to deactivate mentor: %w", err)
	}

	return &mentor, nil
}

func (r *DB) VerifyMentor(mentorID uuid.UUID) (*domain.Mentor, error) {
	query := `
		UPDATE mentors SET
			is_verified = true,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, user_id, company, position, expertise_area, years_experience,
				 availability_type, linkedin_url, website_url, phone, is_verified, is_active,
				 created_at, updated_at`

	var mentor domain.Mentor
	err := r.db.QueryRow(query, mentorID).Scan(
		&mentor.ID,
		&mentor.UserID,
		&mentor.Company,
		&mentor.Position,
		&mentor.ExpertiseArea,
		&mentor.YearsExperience,
		&mentor.AvailabilityType,
		&mentor.LinkedinURL,
		&mentor.WebsiteURL,
		&mentor.Phone,
		&mentor.IsVerified,
		&mentor.IsActive,
		&mentor.CreatedAt,
		&mentor.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("mentor with ID %s not found", mentorID)
		}
		return nil, fmt.Errorf("failed to verify mentor: %w", err)
	}

	return &mentor, nil
}

func (r *DB) GetMentorsByExpertise(expertiseArea string, activeOnly bool) ([]*domain.Mentor, error) {
	query := `
		SELECT id, user_id, company, position, expertise_area, years_experience,
			   availability_type, linkedin_url, website_url, phone, is_verified, is_active,
			   created_at, updated_at
		FROM mentors
		WHERE expertise_area = $1`
	
	args := []interface{}{expertiseArea}
	
	if activeOnly {
		query += " AND is_active = true"
	}
	
	query += " ORDER BY years_experience DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query mentors by expertise: %w", err)
	}
	defer rows.Close()

	var mentors []*domain.Mentor
	for rows.Next() {
		var mentor domain.Mentor
		err := rows.Scan(
			&mentor.ID,
			&mentor.UserID,
			&mentor.Company,
			&mentor.Position,
			&mentor.ExpertiseArea,
			&mentor.YearsExperience,
			&mentor.AvailabilityType,
			&mentor.LinkedinURL,
			&mentor.WebsiteURL,
			&mentor.Phone,
			&mentor.IsVerified,
			&mentor.IsActive,
			&mentor.CreatedAt,
			&mentor.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan mentor: %w", err)
		}
		mentors = append(mentors, &mentor)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return mentors, nil
}
