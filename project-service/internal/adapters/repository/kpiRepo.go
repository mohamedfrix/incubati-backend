package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/moulaybdl/incubAT/project_service/internal/core/domain"
	"github.com/moulaybdl/incubAT/project_service/internal/core/ports"
)

func (r *DB) CreateKPI(kpi *domain.KPI) (*domain.KPI, error) {
	query := `
		INSERT INTO kpis (
			id, project_id, name, description, metric_type, last_updated,
			target_value, current_value, unit, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
		RETURNING id, project_id, name, description, metric_type, last_updated,
				 target_value, current_value, unit, created_at, updated_at, created_by`

	// Generate UUID if not provided
	if kpi.ID == uuid.Nil {
		kpi.ID = uuid.New()
	}

	var createdKPI domain.KPI
	err := r.db.QueryRow(
		query,
		kpi.ID,
		kpi.ProjectID,
		kpi.Name,
		kpi.Description,
		kpi.MetricType,
		kpi.LastUpdated,
		kpi.TargetValue,
		kpi.CurrentValue,
		kpi.Unit,
		kpi.CreatedBy,
	).Scan(
		&createdKPI.ID,
		&createdKPI.ProjectID,
		&createdKPI.Name,
		&createdKPI.Description,
		&createdKPI.MetricType,
		&createdKPI.LastUpdated,
		&createdKPI.TargetValue,
		&createdKPI.CurrentValue,
		&createdKPI.Unit,
		&createdKPI.CreatedAt,
		&createdKPI.UpdatedAt,
		&createdKPI.CreatedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create KPI: %w", err)
	}

	return &createdKPI, nil
}

func (r *DB) UpdateKPI(kpi *domain.KPI) (*domain.KPI, error) {
	query := `
		UPDATE kpis SET
			name = $2,
			description = $3,
			metric_type = $4,
			last_updated = $5,
			target_value = $6,
			current_value = $7,
			unit = $8,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, project_id, name, description, metric_type, last_updated,
				 target_value, current_value, unit, created_at, updated_at, created_by`

	var updatedKPI domain.KPI
	err := r.db.QueryRow(
		query,
		kpi.ID,
		kpi.Name,
		kpi.Description,
		kpi.MetricType,
		kpi.LastUpdated,
		kpi.TargetValue,
		kpi.CurrentValue,
		kpi.Unit,
	).Scan(
		&updatedKPI.ID,
		&updatedKPI.ProjectID,
		&updatedKPI.Name,
		&updatedKPI.Description,
		&updatedKPI.MetricType,
		&updatedKPI.LastUpdated,
		&updatedKPI.TargetValue,
		&updatedKPI.CurrentValue,
		&updatedKPI.Unit,
		&updatedKPI.CreatedAt,
		&updatedKPI.UpdatedAt,
		&updatedKPI.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("KPI with ID %s not found", kpi.ID)
		}
		return nil, fmt.Errorf("failed to update KPI: %w", err)
	}

	return &updatedKPI, nil
}

func (r *DB) DeleteKPI(kpiID uuid.UUID) error {
	// Use DELETE with RETURNING to get KPI info in one query
	query := `DELETE FROM kpis WHERE id = $1 RETURNING name`
	
	var kpiName string
	err := r.db.QueryRow(query, kpiID).Scan(&kpiName)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("KPI with ID %s not found", kpiID)
		}
		return fmt.Errorf("failed to delete KPI: %w", err)
	}

	return nil
}

func (r *DB) GetKPIByID(kpiID uuid.UUID) (*domain.KPI, error) {
	query := `
		SELECT id, project_id, name, description, metric_type, last_updated,
			   target_value, current_value, unit, created_at, updated_at, created_by
		FROM kpis
		WHERE id = $1`

	var kpi domain.KPI
	err := r.db.QueryRow(query, kpiID).Scan(
		&kpi.ID,
		&kpi.ProjectID,
		&kpi.Name,
		&kpi.Description,
		&kpi.MetricType,
		&kpi.LastUpdated,
		&kpi.TargetValue,
		&kpi.CurrentValue,
		&kpi.Unit,
		&kpi.CreatedAt,
		&kpi.UpdatedAt,
		&kpi.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("KPI with ID %s not found", kpiID)
		}
		return nil, fmt.Errorf("failed to get KPI: %w", err)
	}

	return &kpi, nil
}

func (r *DB) GetProjectKPIs(projectID uuid.UUID, filters map[string]interface{}) ([]*domain.KPI, error) {
	// Build the WHERE clause based on filters
	whereConditions := []string{"project_id = $1"}
	args := []interface{}{projectID}
	argIndex := 2

	for key, value := range filters {
		switch key {
		case "metric_type":
			whereConditions = append(whereConditions, fmt.Sprintf("metric_type = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "completed":
			if value.(bool) {
				whereConditions = append(whereConditions, "current_value >= target_value")
			} else {
				whereConditions = append(whereConditions, "current_value < target_value")
			}
		case "over_achieved":
			if value.(bool) {
				whereConditions = append(whereConditions, "current_value > target_value")
			}
		case "under_performing":
			if value.(bool) {
				whereConditions = append(whereConditions, "current_value < target_value * 0.8") // Under 80% of target
			}
		}
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := fmt.Sprintf(`
		SELECT id, project_id, name, description, metric_type, last_updated,
			   target_value, current_value, unit, created_at, updated_at, created_by
		FROM kpis
		WHERE %s
		ORDER BY created_at ASC`, whereClause)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query project KPIs: %w", err)
	}
	defer rows.Close()

	var kpis []*domain.KPI
	for rows.Next() {
		var kpi domain.KPI
		err := rows.Scan(
			&kpi.ID,
			&kpi.ProjectID,
			&kpi.Name,
			&kpi.Description,
			&kpi.MetricType,
			&kpi.LastUpdated,
			&kpi.TargetValue,
			&kpi.CurrentValue,
			&kpi.Unit,
			&kpi.CreatedAt,
			&kpi.UpdatedAt,
			&kpi.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan KPI: %w", err)
		}
		kpis = append(kpis, &kpi)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return kpis, nil
}

func (r *DB) GetAllKPIs(limit, offset int, filters map[string]interface{}) ([]*domain.KPI, int, error) {
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
		case "metric_type":
			whereConditions = append(whereConditions, fmt.Sprintf("metric_type = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "completed":
			if value.(bool) {
				whereConditions = append(whereConditions, "current_value >= target_value")
			} else {
				whereConditions = append(whereConditions, "current_value < target_value")
			}
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM kpis %s", whereClause)
	var totalCount int
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count KPIs: %w", err)
	}

	// Get KPIs with pagination
	query := fmt.Sprintf(`
		SELECT id, project_id, name, description, metric_type, last_updated,
			   target_value, current_value, unit, created_at, updated_at, created_by
		FROM kpis
		%s
		ORDER BY created_at ASC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query KPIs: %w", err)
	}
	defer rows.Close()

	var kpis []*domain.KPI
	for rows.Next() {
		var kpi domain.KPI
		err := rows.Scan(
			&kpi.ID,
			&kpi.ProjectID,
			&kpi.Name,
			&kpi.Description,
			&kpi.MetricType,
			&kpi.LastUpdated,
			&kpi.TargetValue,
			&kpi.CurrentValue,
			&kpi.Unit,
			&kpi.CreatedAt,
			&kpi.UpdatedAt,
			&kpi.CreatedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan KPI: %w", err)
		}
		kpis = append(kpis, &kpi)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return kpis, totalCount, nil
}

func (r *DB) UpdateKPIValue(kpiID uuid.UUID, currentValue float64) (*domain.KPI, error) {
	query := `
		UPDATE kpis SET
			current_value = $2,
			last_updated = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, project_id, name, description, metric_type, last_updated,
				 target_value, current_value, unit, created_at, updated_at, created_by`

	var kpi domain.KPI
	err := r.db.QueryRow(query, kpiID, currentValue).Scan(
		&kpi.ID,
		&kpi.ProjectID,
		&kpi.Name,
		&kpi.Description,
		&kpi.MetricType,
		&kpi.LastUpdated,
		&kpi.TargetValue,
		&kpi.CurrentValue,
		&kpi.Unit,
		&kpi.CreatedAt,
		&kpi.UpdatedAt,
		&kpi.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("KPI with ID %s not found", kpiID)
		}
		return nil, fmt.Errorf("failed to update KPI value: %w", err)
	}

	return &kpi, nil
}

func (r *DB) GetKPIsByMetricType(metricType string, projectID *uuid.UUID) ([]*domain.KPI, error) {
	query := `
		SELECT id, project_id, name, description, metric_type, last_updated,
			   target_value, current_value, unit, created_at, updated_at, created_by
		FROM kpis
		WHERE metric_type = $1`
	
	args := []interface{}{metricType}
	
	if projectID != nil {
		query += " AND project_id = $2"
		args = append(args, *projectID)
	}
	
	query += " ORDER BY created_at ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query KPIs by metric type: %w", err)
	}
	defer rows.Close()

	var kpis []*domain.KPI
	for rows.Next() {
		var kpi domain.KPI
		err := rows.Scan(
			&kpi.ID,
			&kpi.ProjectID,
			&kpi.Name,
			&kpi.Description,
			&kpi.MetricType,
			&kpi.LastUpdated,
			&kpi.TargetValue,
			&kpi.CurrentValue,
			&kpi.Unit,
			&kpi.CreatedAt,
			&kpi.UpdatedAt,
			&kpi.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan KPI: %w", err)
		}
		kpis = append(kpis, &kpi)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return kpis, nil
}

func (r *DB) GetKPIStatistics(projectID *uuid.UUID) (*ports.KPIStatistics, error) {
	query := `
		SELECT 
			COUNT(*) as total_kpis,
			COUNT(CASE WHEN current_value >= target_value THEN 1 END) as completed_kpis,
			COUNT(CASE WHEN current_value > target_value THEN 1 END) as over_achieved_kpis,
			COUNT(CASE WHEN current_value < target_value * 0.8 THEN 1 END) as under_performing_kpis,
			COALESCE(AVG(CASE WHEN target_value > 0 THEN (current_value / target_value) * 100 ELSE 0 END), 0) as average_completion
		FROM kpis`
	
	args := []interface{}{}
	
	if projectID != nil {
		query += " WHERE project_id = $1"
		args = append(args, *projectID)
	}

	var stats ports.KPIStatistics
	var averageCompletion sql.NullFloat64

	err := r.db.QueryRow(query, args...).Scan(
		&stats.TotalKPIs,
		&stats.CompletedKPIs,
		&stats.OverAchievedKPIs,
		&stats.UnderPerformingKPIs,
		&averageCompletion,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get KPI statistics: %w", err)
	}

	// Calculate rates
	if stats.TotalKPIs > 0 {
		stats.AverageCompletion = averageCompletion.Float64
		stats.CompletionRate = float64(stats.CompletedKPIs) / float64(stats.TotalKPIs) * 100
		stats.OverAchievementRate = float64(stats.OverAchievedKPIs) / float64(stats.TotalKPIs) * 100
		stats.UnderPerformanceRate = float64(stats.UnderPerformingKPIs) / float64(stats.TotalKPIs) * 100
	}

	return &stats, nil
}
