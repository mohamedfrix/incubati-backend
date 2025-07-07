package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"moulaybdl/zindy/task_service/internal/core/domain"
	"moulaybdl/zindy/task_service/internal/core/ports"

	"github.com/google/uuid"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) ports.TaskRepository {
	return &TaskRepository{db: db}
}

// Create creates a new task and returns the created task
func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	query := `
		INSERT INTO tasks ( project_id, milestone_id, assigned_to_user_id, created_by_user_id, 
						  title, description, status, priority, due_date, estimated_hours, 
						  actual_hours, parent_task_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, project_id, milestone_id, assigned_to_user_id, created_by_user_id,
				  title, description, status, priority, created_at, updated_at, due_date,
				  completed_at, estimated_hours, actual_hours, parent_task_id
	`

	
	// Set defaults
	task.SetDefaults()
	
	row := r.db.QueryRowContext(ctx, query,
		task.ProjectID,
		task.MilestoneID,
		task.AssignedToUserID,
		task.CreatedByUserID,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.DueDate,
		task.EstimatedHours,
		task.ActualHours,
		task.ParentTaskID,
	)
	
	createdTask := &domain.Task{}
	err := row.Scan(
		&createdTask.ID,
		&createdTask.ProjectID,
		&createdTask.MilestoneID,
		&createdTask.AssignedToUserID,
		&createdTask.CreatedByUserID,
		&createdTask.Title,
		&createdTask.Description,
		&createdTask.Status,
		&createdTask.Priority,
		&createdTask.CreatedAt,
		&createdTask.UpdatedAt,
		&createdTask.DueDate,
		&createdTask.CompletedAt,
		&createdTask.EstimatedHours,
		&createdTask.ActualHours,
		&createdTask.ParentTaskID,
	)
	
	if err != nil {
		return nil, err
	}
	
	return createdTask, nil
}

// GetByID retrieves a task by its ID with related data
func (r *TaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	query := `
		SELECT id, project_id, milestone_id, assigned_to_user_id, created_by_user_id,
			   title, description, status, priority, created_at, updated_at, due_date,
			   completed_at, estimated_hours, actual_hours, parent_task_id
		FROM tasks
		WHERE id = $1
	`
	
	task := &domain.Task{}
	row := r.db.QueryRowContext(ctx, query, id)
	
	err := row.Scan(
		&task.ID,
		&task.ProjectID,
		&task.MilestoneID,
		&task.AssignedToUserID,
		&task.CreatedByUserID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.DueDate,
		&task.CompletedAt,
		&task.EstimatedHours,
		&task.ActualHours,
		&task.ParentTaskID,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrTaskNotFound
		}
		return nil, err
	}
	
	// Load related data
	if err := r.loadRelatedData(ctx, task); err != nil {
		return nil, err
	}
	
	return task, nil
}

// Update updates an existing task
func (r *TaskRepository) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateTaskRequest) (*domain.Task, error) {
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1
	
	if req.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, *req.Title)
		argIndex++
	}
	
	if req.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, req.Description)
		argIndex++
	}
	
	if req.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *req.Status)
		argIndex++
	}
	
	if req.Priority != nil {
		setParts = append(setParts, fmt.Sprintf("priority = $%d", argIndex))
		args = append(args, *req.Priority)
		argIndex++
	}
	
	if req.DueDate != nil {
		setParts = append(setParts, fmt.Sprintf("due_date = $%d", argIndex))
		args = append(args, req.DueDate)
		argIndex++
	}
	
	if req.EstimatedHours != nil {
		setParts = append(setParts, fmt.Sprintf("estimated_hours = $%d", argIndex))
		args = append(args, req.EstimatedHours)
		argIndex++
	}
	
	if req.ActualHours != nil {
		setParts = append(setParts, fmt.Sprintf("actual_hours = $%d", argIndex))
		args = append(args, req.ActualHours)
		argIndex++
	}
	
	if req.AssignedToUserID != nil {
		setParts = append(setParts, fmt.Sprintf("assigned_to_user_id = $%d", argIndex))
		args = append(args, req.AssignedToUserID)
		argIndex++
	}
	
	if req.ProjectID != nil {
		setParts = append(setParts, fmt.Sprintf("project_id = $%d", argIndex))
		args = append(args, *req.ProjectID)
		argIndex++
	}
	
	if req.MilestoneID != nil {
		setParts = append(setParts, fmt.Sprintf("milestone_id = $%d", argIndex))
		args = append(args, *req.MilestoneID)
		argIndex++
	}
	
	if req.ParentTaskID != nil {
		setParts = append(setParts, fmt.Sprintf("parent_task_id = $%d", argIndex))
		args = append(args, req.ParentTaskID)
		argIndex++
	}
	
	if len(setParts) == 0 {
		return r.GetByID(ctx, id)
	}
	
	// Always update the updated_at field
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++
	
	// Add the WHERE clause parameter
	args = append(args, id)
	
	query := fmt.Sprintf(`
		UPDATE tasks 
		SET %s
		WHERE id = $%d
	`, strings.Join(setParts, ", "), argIndex)
	
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	
	return r.GetByID(ctx, id)
}

// Delete deletes a task by its ID
func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if task has subtasks
	hasSubtasks, err := r.HasSubtasks(ctx, id)
	if err != nil {
		return err
	}
	if hasSubtasks {
		return domain.ErrTaskHasSubtasks
	}
	
	query := `DELETE FROM tasks WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return domain.ErrTaskNotFound
	}
	
	return nil
}

// GetTasksByProject retrieves all tasks for a specific project with filtering
func (r *TaskRepository) GetTasksByProject(ctx context.Context, projectID uuid.UUID, filter *domain.TaskFilterRequest) ([]domain.Task, *domain.PaginationInfo, error) {
	baseQuery := `
		SELECT id, project_id, milestone_id, assigned_to_user_id, created_by_user_id,
			   title, description, status, priority, created_at, updated_at, due_date,
			   completed_at, estimated_hours, actual_hours, parent_task_id
		FROM tasks
		WHERE project_id = $1
	`
	
	countQuery := `SELECT COUNT(*) FROM tasks WHERE project_id = $1`
	
	whereClause, args := r.buildWhereClause(filter, []interface{}{projectID})
	if whereClause != "" {
		baseQuery += " AND " + whereClause
		countQuery += " AND " + whereClause
	}
	
	// Add ordering
	orderBy := "created_at DESC"
	if filter != nil && filter.OrderBy != nil {
		orderBy = r.buildOrderBy(*filter.OrderBy)
	}
	baseQuery += " ORDER BY " + orderBy
	
	// Add pagination
	page, pageSize := r.getPaginationParams(filter)
	offset := (page - 1) * pageSize
	baseQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
	
	// Get total count
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, nil, err
	}
	
	// Get tasks
	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	
	tasks, err := r.scanTasks(rows)
	if err != nil {
		return nil, nil, err
	}
	
	pagination := &domain.PaginationInfo{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: (totalCount + pageSize - 1) / pageSize,
	}
	
	return tasks, pagination, nil
}

// GetTasksByUser retrieves all tasks assigned to a specific user
func (r *TaskRepository) GetTasksByUser(ctx context.Context, userID uuid.UUID, filter *domain.TaskFilterRequest) ([]domain.Task, *domain.PaginationInfo, error) {
	baseQuery := `
		SELECT id, project_id, milestone_id, assigned_to_user_id, created_by_user_id,
			   title, description, status, priority, created_at, updated_at, due_date,
			   completed_at, estimated_hours, actual_hours, parent_task_id
		FROM tasks
		WHERE assigned_to_user_id = $1
	`
	
	countQuery := `SELECT COUNT(*) FROM tasks WHERE assigned_to_user_id = $1`
	
	whereClause, args := r.buildWhereClause(filter, []interface{}{userID})
	if whereClause != "" {
		baseQuery += " AND " + whereClause
		countQuery += " AND " + whereClause
	}
	
	// Add ordering
	orderBy := "created_at DESC"
	if filter != nil && filter.OrderBy != nil {
		orderBy = r.buildOrderBy(*filter.OrderBy)
	}
	baseQuery += " ORDER BY " + orderBy
	
	// Add pagination
	page, pageSize := r.getPaginationParams(filter)
	offset := (page - 1) * pageSize
	baseQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
	
	// Get total count
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, nil, err
	}
	
	// Get tasks
	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	
	tasks, err := r.scanTasks(rows)
	if err != nil {
		return nil, nil, err
	}
	
	pagination := &domain.PaginationInfo{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: (totalCount + pageSize - 1) / pageSize,
	}
	
	return tasks, pagination, nil
}

// AssignTask assigns a task to a user
func (r *TaskRepository) AssignTask(ctx context.Context, taskID uuid.UUID, req *domain.AssignTaskRequest) (*domain.Task, error) {
	query := `
		UPDATE tasks 
		SET assigned_to_user_id = $1, updated_at = $2
		WHERE id = $3
	`
	
	_, err := r.db.ExecContext(ctx, query, req.AssignedToUserID, time.Now(), taskID)
	if err != nil {
		return nil, err
	}
	
	return r.GetByID(ctx, taskID)
}

// UnassignTask removes assignment from a task
func (r *TaskRepository) UnassignTask(ctx context.Context, taskID uuid.UUID) (*domain.Task, error) {
	query := `
		UPDATE tasks 
		SET assigned_to_user_id = NULL, updated_at = $1
		WHERE id = $2
	`
	
	_, err := r.db.ExecContext(ctx, query, time.Now(), taskID)
	if err != nil {
		return nil, err
	}
	
	return r.GetByID(ctx, taskID)
}

// ChangeStatus updates only the status of a task
func (r *TaskRepository) ChangeStatus(ctx context.Context, taskID uuid.UUID, req *domain.ChangeStatusRequest) (*domain.Task, error) {
	var complete_time *time.Time 
	if req.Status == domain.TaskStatusDone {
		now := time.Now()
		complete_time = &now
	} else {
		complete_time = nil
	}
	query := `
		UPDATE tasks 
		SET status = $1, updated_at = $2, completed_at = $4
		WHERE id = $3
	`
	
	result, err := r.db.ExecContext(ctx, query, req.Status, time.Now(), taskID, complete_time)
	if err != nil {
		return nil, err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	
	if rowsAffected == 0 {
		return nil, domain.ErrTaskNotFound
	}
	
	return r.GetByID(ctx, taskID)
}

// BulkUpdate updates multiple tasks at once
func (r *TaskRepository) BulkUpdate(ctx context.Context, req *domain.BulkUpdateTasksRequest) ([]domain.Task, error) {
	if len(req.TaskIDs) == 0 {
		return nil, domain.ErrNoTasksSelected
	}
	
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1
	
	if req.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *req.Status)
		argIndex++
	}
	
	if req.Priority != nil {
		setParts = append(setParts, fmt.Sprintf("priority = $%d", argIndex))
		args = append(args, *req.Priority)
		argIndex++
	}
	
	if req.AssignedToUserID != nil {
		setParts = append(setParts, fmt.Sprintf("assigned_to_user_id = $%d", argIndex))
		args = append(args, req.AssignedToUserID)
		argIndex++
	}
	
	if req.DueDate != nil {
		setParts = append(setParts, fmt.Sprintf("due_date = $%d", argIndex))
		args = append(args, req.DueDate)
		argIndex++
	}
	
	if len(setParts) == 0 {
		return nil, domain.ErrNoUpdateFieldsProvided
	}
	
	// Always update the updated_at field
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Always update the completed_at field if status is set to Done
	if req.Status != nil && *req.Status == domain.TaskStatusDone {
		setParts = append(setParts, fmt.Sprintf("completed_at = $%d", argIndex))
		args = append(args, time.Now())
		argIndex++
	} 

	
	// Build IN clause for task IDs
	placeholders := make([]string, len(req.TaskIDs))
	for i, taskID := range req.TaskIDs {
		placeholders[i] = fmt.Sprintf("$%d", argIndex)
		args = append(args, taskID)
		argIndex++
	}
	
	query := fmt.Sprintf(`
		UPDATE tasks 
		SET %s
		WHERE id IN (%s)
	`, strings.Join(setParts, ", "), strings.Join(placeholders, ", "))
	
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	
	// Return updated tasks
	var tasks []domain.Task
	for _, taskID := range req.TaskIDs {
		task, err := r.GetByID(ctx, taskID)
		if err != nil {
			continue // Skip tasks that might have been deleted
		}
		tasks = append(tasks, *task)
	}
	
	return tasks, nil
}

// GetTaskStatistics returns detailed statistics for a task
func (r *TaskRepository) GetTaskStatistics(ctx context.Context, taskID uuid.UUID) (*domain.TaskStatistics, error) {
	task, err := r.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	
	stats := &domain.TaskStatistics{
		Task: *task,
	}
	
	// Get subtask statistics
	subtaskStats, err := r.getSubtaskStatistics(ctx, taskID)
	if err != nil {
		return nil, err
	}
	stats.SubtaskStats = *subtaskStats
	
	// Get comment statistics
	commentStats, err := r.getCommentStatistics(ctx, taskID)
	if err != nil {
		return nil, err
	}
	stats.CommentStats = *commentStats
	
	// Get attachment statistics
	attachmentStats, err := r.getAttachmentStatistics(ctx, taskID)
	if err != nil {
		return nil, err
	}
	stats.AttachmentStats = *attachmentStats
	
	// Calculate time tracking
	stats.TimeTracking = r.calculateTimeTracking(task)
	
	return stats, nil
}

// GetTaskSummary returns summary statistics for tasks
func (r *TaskRepository) GetTaskSummary(ctx context.Context, projectID *uuid.UUID, userID *uuid.UUID) (*domain.TaskSummary, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIndex := 1
	
	if projectID != nil {
		whereClause += fmt.Sprintf(" AND project_id = $%d", argIndex)
		args = append(args, *projectID)
		argIndex++
	}
	
	if userID != nil {
		whereClause += fmt.Sprintf(" AND assigned_to_user_id = $%d", argIndex)
		args = append(args, *userID)
		argIndex++
	}
	
	// Get total count
	totalQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks %s", whereClause)
	var totalTasks int
	err := r.db.QueryRowContext(ctx, totalQuery, args...).Scan(&totalTasks)
	if err != nil {
		return nil, err
	}
	
	// Get status breakdown
	statusQuery := fmt.Sprintf(`
		SELECT status, COUNT(*) 
		FROM tasks %s 
		GROUP BY status
	`, whereClause)
	
	rows, err := r.db.QueryContext(ctx, statusQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	byStatus := make(map[domain.TaskStatus]int)
	for rows.Next() {
		var status domain.TaskStatus
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		byStatus[status] = count
	}
	
	// Get priority breakdown
	priorityQuery := fmt.Sprintf(`
		SELECT priority, COUNT(*) 
		FROM tasks %s 
		GROUP BY priority
	`, whereClause)
	
	rows, err = r.db.QueryContext(ctx, priorityQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	byPriority := make(map[domain.TaskPriority]int)
	for rows.Next() {
		var priority domain.TaskPriority
		var count int
		if err := rows.Scan(&priority, &count); err != nil {
			return nil, err
		}
		byPriority[priority] = count
	}
	
	return &domain.TaskSummary{
		TotalTasks: totalTasks,
		ByStatus:   byStatus,
		ByPriority: byPriority,
	}, nil
}

// GetSubtasks retrieves subtasks for a parent task
func (r *TaskRepository) GetSubtasks(ctx context.Context, parentTaskID uuid.UUID, includeSubtasks bool) ([]domain.Task, error) {
	query := `
		SELECT id, project_id, milestone_id, assigned_to_user_id, created_by_user_id,
			   title, description, status, priority, created_at, updated_at, due_date,
			   completed_at, estimated_hours, actual_hours, parent_task_id
		FROM tasks
		WHERE parent_task_id = $1
		ORDER BY created_at ASC
	`
	
	rows, err := r.db.QueryContext(ctx, query, parentTaskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	tasks, err := r.scanTasks(rows)
	if err != nil {
		return nil, err
	}
	
	// Recursively load subtasks if requested
	if includeSubtasks {
		for i := range tasks {
			subtasks, err := r.GetSubtasks(ctx, tasks[i].ID, true)
			if err != nil {
				return nil, err
			}
			tasks[i].Subtasks = subtasks
		}
	}
	
	return tasks, nil
}

// HasSubtasks checks if a task has subtasks
func (r *TaskRepository) HasSubtasks(ctx context.Context, taskID uuid.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM tasks WHERE parent_task_id = $1`
	
	var count int
	err := r.db.QueryRowContext(ctx, query, taskID).Scan(&count)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// ValidateParentTask validates parent-child relationship
func (r *TaskRepository) ValidateParentTask(ctx context.Context, taskID uuid.UUID, parentTaskID *uuid.UUID) error {
	if parentTaskID == nil {
		return nil
	}
	
	// Check self-reference
	if *parentTaskID == taskID {
		return domain.ErrSelfReferenceTask
	}
	
	// Check if parent exists
	_, err := r.GetByID(ctx, *parentTaskID)
	if err != nil {
		return err
	}
	
	// Check for circular reference (simplified check)
	query := `
		WITH RECURSIVE task_hierarchy AS (
			SELECT id, parent_task_id, 1 as level
			FROM tasks 
			WHERE id = $1
			
			UNION ALL
			
			SELECT t.id, t.parent_task_id, th.level + 1
			FROM tasks t
			INNER JOIN task_hierarchy th ON t.id = th.parent_task_id
			WHERE th.level < 10
		)
		SELECT COUNT(*) FROM task_hierarchy WHERE id = $2
	`
	
	var count int
	err = r.db.QueryRowContext(ctx, query, *parentTaskID, taskID).Scan(&count)
	if err != nil {
		return err
	}
	
	if count > 0 {
		return domain.ErrCircularReference
	}
	
	return nil
}

// Helper methods

func (r *TaskRepository) scanTasks(rows *sql.Rows) ([]domain.Task, error) {
	var tasks []domain.Task
	
	for rows.Next() {
		var task domain.Task
		err := rows.Scan(
			&task.ID,
			&task.ProjectID,
			&task.MilestoneID,
			&task.AssignedToUserID,
			&task.CreatedByUserID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.DueDate,
			&task.CompletedAt,
			&task.EstimatedHours,
			&task.ActualHours,
			&task.ParentTaskID,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return tasks, nil
}

func (r *TaskRepository) loadRelatedData(ctx context.Context, task *domain.Task) error {
	// Load comments
	commentQuery := `
		SELECT id, task_id, author_user_id, content, created_at, updated_at
		FROM task_comments
		WHERE task_id = $1
		ORDER BY created_at ASC
	`
	
	rows, err := r.db.QueryContext(ctx, commentQuery, task.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	for rows.Next() {
		var comment domain.TaskComment
		err := rows.Scan(
			&comment.ID,
			&comment.TaskID,
			&comment.AuthorUserID,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return err
		}
		task.Comments = append(task.Comments, comment)
	}
	
	// Load attachments
	attachmentQuery := `
		SELECT id, task_id, uploaded_by_user_id, file_name, file_url, file_size, mime_type, uploaded_at
		FROM task_attachments
		WHERE task_id = $1
		ORDER BY uploaded_at DESC
	`
	
	rows, err = r.db.QueryContext(ctx, attachmentQuery, task.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	for rows.Next() {
		var attachment domain.TaskAttachment
		err := rows.Scan(
			&attachment.ID,
			&attachment.TaskID,
			&attachment.UploadedByUserID,
			&attachment.FileName,
			&attachment.FileURL,
			&attachment.FileSize,
			&attachment.MimeType,
			&attachment.UploadedAt,
		)
		if err != nil {
			return err
		}
		task.Attachments = append(task.Attachments, attachment)
	}
	
	// Load subtasks
	subtasks, err := r.GetSubtasks(ctx, task.ID, false)
	if err != nil {
		return err
	}
	task.Subtasks = subtasks
	
	return nil
}

func (r *TaskRepository) buildWhereClause(filter *domain.TaskFilterRequest, args []interface{}) (string, []interface{}) {
	if filter == nil {
		return "", args
	}
	
	var conditions []string
	argIndex := len(args) + 1
	
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}
	
	if filter.Priority != nil {
		conditions = append(conditions, fmt.Sprintf("priority = $%d", argIndex))
		args = append(args, *filter.Priority)
		argIndex++
	}
	
	if filter.AssignedToUserID != nil {
		conditions = append(conditions, fmt.Sprintf("assigned_to_user_id = $%d", argIndex))
		args = append(args, *filter.AssignedToUserID)
		argIndex++
	}
	
	if filter.ProjectID != nil {
		conditions = append(conditions, fmt.Sprintf("project_id = $%d", argIndex))
		args = append(args, *filter.ProjectID)
		argIndex++
	}
	
	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
		searchTerm := "%" + *filter.Search + "%"
		args = append(args, searchTerm)
		argIndex++
	}
	
	if filter.OverdueOnly != nil && *filter.OverdueOnly {
		conditions = append(conditions, "due_date < NOW() AND status != 'done'")
	}
	
	if filter.DueSoon != nil && *filter.DueSoon > 0 {
		conditions = append(conditions, fmt.Sprintf("due_date <= NOW() + INTERVAL '%d days'", *filter.DueSoon))
	}
	
	if filter.IncludeSubtasks != nil && !*filter.IncludeSubtasks {
		conditions = append(conditions, "parent_task_id IS NULL")
	}
	
	return strings.Join(conditions, " AND "), args
}

func (r *TaskRepository) buildOrderBy(orderBy string) string {
	allowedFields := map[string]bool{
		"created_at":       true,
		"-created_at":      true,
		"updated_at":       true,
		"-updated_at":      true,
		"title":            true,
		"-title":           true,
		"status":           true,
		"-status":          true,
		"priority":         true,
		"-priority":        true,
		"due_date":         true,
		"-due_date":        true,
	}
	
	if !allowedFields[orderBy] {
		return "created_at DESC"
	}
	
	if strings.HasPrefix(orderBy, "-") {
		return strings.TrimPrefix(orderBy, "-") + " DESC"
	}
	
	return orderBy + " ASC"
}

func (r *TaskRepository) getPaginationParams(filter *domain.TaskFilterRequest) (page, pageSize int) {
	page = 1
	pageSize = 10
	
	if filter != nil {
		if filter.Page != nil && *filter.Page > 0 {
			page = *filter.Page
		}
		if filter.PageSize != nil && *filter.PageSize > 0 && *filter.PageSize <= 100 {
			pageSize = *filter.PageSize
		}
	}
	
	return page, pageSize
}

func (r *TaskRepository) getSubtaskStatistics(ctx context.Context, taskID uuid.UUID) (*domain.SubtaskStatistics, error) {
	query := `
		SELECT status, COUNT(*)
		FROM tasks
		WHERE parent_task_id = $1
		GROUP BY status
	`
	
	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	stats := &domain.SubtaskStatistics{
		ByStatus: make(map[domain.TaskStatus]int),
	}
	
	for rows.Next() {
		var status domain.TaskStatus
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		
		stats.Total += count
		stats.ByStatus[status] = count
		
		if status == domain.TaskStatusDone {
			stats.Completed += count
		} else if status == domain.TaskStatusInProgress {
			stats.InProgress += count
		}
	}
	
	if stats.Total > 0 {
		stats.Completion = float64(stats.Completed) / float64(stats.Total) * 100
	}
	
	return stats, nil
}

func (r *TaskRepository) getCommentStatistics(ctx context.Context, taskID uuid.UUID) (*domain.CommentStatistics, error) {
	query := `
		SELECT COUNT(*), COUNT(DISTINCT author_user_id), MAX(created_at)
		FROM task_comments
		WHERE task_id = $1
	`
	
	stats := &domain.CommentStatistics{
		AuthorActivity: make(map[string]int),
	}
	
	var lastCommentAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, taskID).Scan(&stats.Total, &stats.UniqueAuthors, &lastCommentAt)
	if err != nil {
		return nil, err
	}
	
	if lastCommentAt.Valid {
		stats.LastCommentAt = &lastCommentAt.Time
	}
	
	return stats, nil
}

func (r *TaskRepository) getAttachmentStatistics(ctx context.Context, taskID uuid.UUID) (*domain.AttachmentStatistics, error) {
	query := `
		SELECT COUNT(*), COALESCE(SUM(file_size), 0), MAX(uploaded_at)
		FROM task_attachments
		WHERE task_id = $1
	`
	
	stats := &domain.AttachmentStatistics{
		ByType: make(map[string]int),
	}
	
	var lastUpload sql.NullTime
	err := r.db.QueryRowContext(ctx, query, taskID).Scan(&stats.Total, &stats.TotalSize, &lastUpload)
	if err != nil {
		return nil, err
	}
	
	if lastUpload.Valid {
		stats.LastUpload = &lastUpload.Time
	}
	
	stats.TotalSizeMB = float64(stats.TotalSize) / (1024 * 1024)
	
	return stats, nil
}

func (r *TaskRepository) calculateTimeTracking(task *domain.Task) domain.TimeTrackingInfo {
	tracking := domain.TimeTrackingInfo{
		EstimatedHours: task.EstimatedHours,
		ActualHours:    task.ActualHours,
	}
	
	if task.EstimatedHours != nil && task.ActualHours != nil {
		remaining := *task.EstimatedHours - *task.ActualHours
		tracking.RemainingHours = &remaining
		tracking.IsOverEstimate = *task.ActualHours > *task.EstimatedHours
		
		if *task.EstimatedHours > 0 {
			ratio := *task.ActualHours / *task.EstimatedHours
			tracking.CompletionRatio = &ratio
		}
	}
	
	return tracking
}
