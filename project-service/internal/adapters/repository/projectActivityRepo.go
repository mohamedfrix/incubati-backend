package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
)

func (r *DB) CreateActivity(activity *domain.ProjectActivity) (*domain.ProjectActivity, error) {
	query := `
		INSERT INTO project_activities (
			id, project_id, activity_type, description, metadata,
			user_id, related_object_id, related_object_type
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
		RETURNING id, project_id, activity_type, description, metadata,
				 user_id, related_object_id, related_object_type, created_at`

	// Generate UUID if not provided
	if activity.ID == uuid.Nil {
		activity.ID = uuid.New()
	}

	// Convert metadata to JSON
	var metadataJSON []byte
	var err error
	if activity.Metadata != nil {
		metadataJSON, err = json.Marshal(activity.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
	} else {
		metadataJSON = []byte("{}")
	}

	var createdActivity domain.ProjectActivity
	var metadataStr string
	err = r.db.QueryRow(
		query,
		activity.ID,
		activity.ProjectID,
		activity.ActivityType,
		activity.Description,
		metadataJSON,
		activity.UserID,
		activity.RelatedObjectID,
		activity.RelatedObjectType,
	).Scan(
		&createdActivity.ID,
		&createdActivity.ProjectID,
		&createdActivity.ActivityType,
		&createdActivity.Description,
		&metadataStr,
		&createdActivity.UserID,
		&createdActivity.RelatedObjectID,
		&createdActivity.RelatedObjectType,
		&createdActivity.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create activity: %w", err)
	}

	// Parse metadata back
	if metadataStr != "" {
		err = json.Unmarshal([]byte(metadataStr), &createdActivity.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &createdActivity, nil
}

func (r *DB) GetActivityByID(activityID uuid.UUID) (*domain.ProjectActivity, error) {
	query := `
		SELECT id, project_id, activity_type, description, metadata,
			   user_id, related_object_id, related_object_type, created_at
		FROM project_activities
		WHERE id = $1`

	var activity domain.ProjectActivity
	var metadataStr string
	err := r.db.QueryRow(query, activityID).Scan(
		&activity.ID,
		&activity.ProjectID,
		&activity.ActivityType,
		&activity.Description,
		&metadataStr,
		&activity.UserID,
		&activity.RelatedObjectID,
		&activity.RelatedObjectType,
		&activity.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("activity with ID %s not found", activityID)
		}
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	// Parse metadata
	if metadataStr != "" {
		err = json.Unmarshal([]byte(metadataStr), &activity.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &activity, nil
}

func (r *DB) GetProjectActivities(projectID uuid.UUID, filters map[string]interface{}) ([]*domain.ProjectActivity, error) {
	// Build the WHERE clause based on filters
	whereConditions := []string{"project_id = $1"}
	args := []interface{}{projectID}
	argIndex := 2

	for key, value := range filters {
		switch key {
		case "activity_type":
			whereConditions = append(whereConditions, fmt.Sprintf("activity_type = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "user_id":
			whereConditions = append(whereConditions, fmt.Sprintf("user_id = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "related_object_type":
			whereConditions = append(whereConditions, fmt.Sprintf("related_object_type = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "related_object_id":
			whereConditions = append(whereConditions, fmt.Sprintf("related_object_id = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "recent_days":
			days := value.(int)
			whereConditions = append(whereConditions, fmt.Sprintf("created_at >= CURRENT_DATE - INTERVAL '%d days'", days))
		}
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := fmt.Sprintf(`
		SELECT id, project_id, activity_type, description, metadata,
			   user_id, related_object_id, related_object_type, created_at
		FROM project_activities
		WHERE %s
		ORDER BY created_at DESC`, whereClause)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query project activities: %w", err)
	}
	defer rows.Close()

	var activities []*domain.ProjectActivity
	for rows.Next() {
		var activity domain.ProjectActivity
		var metadataStr string
		err := rows.Scan(
			&activity.ID,
			&activity.ProjectID,
			&activity.ActivityType,
			&activity.Description,
			&metadataStr,
			&activity.UserID,
			&activity.RelatedObjectID,
			&activity.RelatedObjectType,
			&activity.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}

		// Parse metadata
		if metadataStr != "" {
			err = json.Unmarshal([]byte(metadataStr), &activity.Metadata)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		activities = append(activities, &activity)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return activities, nil
}

func (r *DB) GetAllActivities(limit, offset int, filters map[string]interface{}) ([]*domain.ProjectActivity, int, error) {
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
		case "activity_type":
			whereConditions = append(whereConditions, fmt.Sprintf("activity_type = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "user_id":
			whereConditions = append(whereConditions, fmt.Sprintf("user_id = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "related_object_type":
			whereConditions = append(whereConditions, fmt.Sprintf("related_object_type = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "search":
			whereConditions = append(whereConditions, fmt.Sprintf("description ILIKE $%d", argIndex))
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM project_activities %s", whereClause)
	var totalCount int
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count activities: %w", err)
	}

	// Get activities with pagination
	query := fmt.Sprintf(`
		SELECT id, project_id, activity_type, description, metadata,
			   user_id, related_object_id, related_object_type, created_at
		FROM project_activities
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query activities: %w", err)
	}
	defer rows.Close()

	var activities []*domain.ProjectActivity
	for rows.Next() {
		var activity domain.ProjectActivity
		var metadataStr string
		err := rows.Scan(
			&activity.ID,
			&activity.ProjectID,
			&activity.ActivityType,
			&activity.Description,
			&metadataStr,
			&activity.UserID,
			&activity.RelatedObjectID,
			&activity.RelatedObjectType,
			&activity.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan activity: %w", err)
		}

		// Parse metadata
		if metadataStr != "" {
			err = json.Unmarshal([]byte(metadataStr), &activity.Metadata)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		activities = append(activities, &activity)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return activities, totalCount, nil
}

func (r *DB) GetActivitiesByUser(userID uuid.UUID, projectID *uuid.UUID) ([]*domain.ProjectActivity, error) {
	query := `
		SELECT id, project_id, activity_type, description, metadata,
			   user_id, related_object_id, related_object_type, created_at
		FROM project_activities
		WHERE user_id = $1`
	
	args := []interface{}{userID}
	
	if projectID != nil {
		query += " AND project_id = $2"
		args = append(args, *projectID)
	}
	
	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query activities by user: %w", err)
	}
	defer rows.Close()

	var activities []*domain.ProjectActivity
	for rows.Next() {
		var activity domain.ProjectActivity
		var metadataStr string
		err := rows.Scan(
			&activity.ID,
			&activity.ProjectID,
			&activity.ActivityType,
			&activity.Description,
			&metadataStr,
			&activity.UserID,
			&activity.RelatedObjectID,
			&activity.RelatedObjectType,
			&activity.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}

		// Parse metadata
		if metadataStr != "" {
			err = json.Unmarshal([]byte(metadataStr), &activity.Metadata)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		activities = append(activities, &activity)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return activities, nil
}

func (r *DB) GetActivitiesByType(activityType string, projectID *uuid.UUID) ([]*domain.ProjectActivity, error) {
	query := `
		SELECT id, project_id, activity_type, description, metadata,
			   user_id, related_object_id, related_object_type, created_at
		FROM project_activities
		WHERE activity_type = $1`
	
	args := []interface{}{activityType}
	
	if projectID != nil {
		query += " AND project_id = $2"
		args = append(args, *projectID)
	}
	
	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query activities by type: %w", err)
	}
	defer rows.Close()

	var activities []*domain.ProjectActivity
	for rows.Next() {
		var activity domain.ProjectActivity
		var metadataStr string
		err := rows.Scan(
			&activity.ID,
			&activity.ProjectID,
			&activity.ActivityType,
			&activity.Description,
			&metadataStr,
			&activity.UserID,
			&activity.RelatedObjectID,
			&activity.RelatedObjectType,
			&activity.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}

		// Parse metadata
		if metadataStr != "" {
			err = json.Unmarshal([]byte(metadataStr), &activity.Metadata)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		activities = append(activities, &activity)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return activities, nil
}

func (r *DB) GetRecentActivities(projectID *uuid.UUID, days int) ([]*domain.ProjectActivity, error) {
	query := `
		SELECT id, project_id, activity_type, description, metadata,
			   user_id, related_object_id, related_object_type, created_at
		FROM project_activities
		WHERE created_at >= CURRENT_DATE - INTERVAL '%d days'`
	
	args := []interface{}{}
	
	if projectID != nil {
		query += " AND project_id = $1"
		args = append(args, *projectID)
	}
	
	query += " ORDER BY created_at DESC"

	formattedQuery := fmt.Sprintf(query, days)

	rows, err := r.db.Query(formattedQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent activities: %w", err)
	}
	defer rows.Close()

	var activities []*domain.ProjectActivity
	for rows.Next() {
		var activity domain.ProjectActivity
		var metadataStr string
		err := rows.Scan(
			&activity.ID,
			&activity.ProjectID,
			&activity.ActivityType,
			&activity.Description,
			&metadataStr,
			&activity.UserID,
			&activity.RelatedObjectID,
			&activity.RelatedObjectType,
			&activity.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recent activity: %w", err)
		}

		// Parse metadata
		if metadataStr != "" {
			err = json.Unmarshal([]byte(metadataStr), &activity.Metadata)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		activities = append(activities, &activity)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return activities, nil
}

func (r *DB) DeleteActivity(activityID uuid.UUID) error {
	// Use DELETE with RETURNING to get activity info in one query
	query := `DELETE FROM project_activities WHERE id = $1 RETURNING activity_type`
	
	var activityType string
	err := r.db.QueryRow(query, activityID).Scan(&activityType)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("activity with ID %s not found", activityID)
		}
		return fmt.Errorf("failed to delete activity: %w", err)
	}

	return nil
}

func (r *DB) GetActivityStatistics(projectID *uuid.UUID) (*ports.ActivityStatistics, error) {
	query := `
		SELECT 
			COUNT(*) as total_activities,
			COUNT(CASE WHEN created_at >= CURRENT_DATE - INTERVAL '30 days' THEN 1 END) as recent_activities
		FROM project_activities`
	
	args := []interface{}{}
	
	if projectID != nil {
		query += " WHERE project_id = $1"
		args = append(args, *projectID)
	}

	var stats ports.ActivityStatistics

	err := r.db.QueryRow(query, args...).Scan(
		&stats.TotalActivities,
		&stats.RecentActivities30Days,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get activity statistics: %w", err)
	}

	// Get activities by type
	typeQuery := `
		SELECT activity_type, COUNT(*) 
		FROM project_activities`
	
	if projectID != nil {
		typeQuery += " WHERE project_id = $1"
	}
	
	typeQuery += " GROUP BY activity_type"

	typeRows, err := r.db.Query(typeQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity type statistics: %w", err)
	}
	defer typeRows.Close()

	stats.ActivitiesByType = make(map[string]int)
	for typeRows.Next() {
		var activityType string
		var count int
		err := typeRows.Scan(&activityType, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity type: %w", err)
		}
		stats.ActivitiesByType[activityType] = count
	}

	// Get activities by user
	userQuery := `
		SELECT user_id::text, COUNT(*) 
		FROM project_activities`
	
	if projectID != nil {
		userQuery += " WHERE project_id = $1"
	}
	
	userQuery += " GROUP BY user_id ORDER BY COUNT(*) DESC"

	userRows, err := r.db.Query(userQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get user statistics: %w", err)
	}
	defer userRows.Close()

	stats.ActivitiesByUser = make(map[string]int)
	stats.MostActiveUsers = []ports.UserStats{}
	for userRows.Next() {
		var userID string
		var count int
		err := userRows.Scan(&userID, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		stats.ActivitiesByUser[userID] = count
		stats.MostActiveUsers = append(stats.MostActiveUsers, ports.UserStats{
			UserID:        userID,
			ActivityCount: count,
		})
	}

	// Get activity trend (last 7 days)
	trendQuery := `
		SELECT DATE(created_at) as activity_date, COUNT(*) as activity_count
		FROM project_activities
		WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'`
	
	if projectID != nil {
		trendQuery += " AND project_id = $1"
	}
	
	trendQuery += " GROUP BY DATE(created_at) ORDER BY activity_date"

	trendRows, err := r.db.Query(trendQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity trend: %w", err)
	}
	defer trendRows.Close()

	stats.ActivityTrend = []ports.DailyStats{}
	for trendRows.Next() {
		var date string
		var count int
		err := trendRows.Scan(&date, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan activity trend: %w", err)
		}
		stats.ActivityTrend = append(stats.ActivityTrend, ports.DailyStats{
			Date:  date,
			Count: count,
		})
	}

	return &stats, nil
}
