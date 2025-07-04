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
		INSERT INTO tasks (id, title, description, status, priority, due_date, completed_at, assignee_id, project_id, milestone_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, title, description, status, priority, due_date, completed_at, assignee_id, project_id, milestone_id, created_at, updated_at
	`
	
	now := time.Now()
	task.ID = uuid.New()
	task.CreatedAt = now
	task.UpdatedAt = now
	
	// Set defaults
	task.SetDefaults()
	
	row := r.db.QueryRowContext(ctx, query,
		task.ID,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.DueDate,
		task.CompletedAt,
		task.AssigneeID,
		task.ProjectID,
		task.MilestoneID,
		task.CreatedAt,
		task.UpdatedAt,
	)
	
	createdTask := &domain.Task{}
	err := row.Scan(
		&createdTask.ID,
		&createdTask.Title,
		&createdTask.Description,
		&createdTask.Status,
		&createdTask.Priority,
		&createdTask.DueDate,
		&createdTask.CompletedAt,
		&createdTask.AssigneeID,
		&createdTask.ProjectID,
		&createdTask.MilestoneID,
		&createdTask.CreatedAt,
		&createdTask.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return createdTask, nil
}

// GetByID retrieves a task by its ID
func (r *TaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	query := `
		SELECT id, title, description, status, priority, due_date, completed_at, assignee_id, project_id, milestone_id, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`
	
	task := &domain.Task{}
	row := r.db.QueryRowContext(ctx, query, id)
	
	err := row.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.DueDate,
		&task.CompletedAt,
		&task.AssigneeID,
		&task.ProjectID,
		&task.MilestoneID,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrTaskNotFound
		}
		return nil, err
	}
	
	return task, nil
}

// Update updates an existing task
func (r *TaskRepository) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateTaskRequest) (*domain.Task, error) {
	// Build dynamic query based on provided fields
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
	
	if req.CompletedAt != nil {
		setParts = append(setParts, fmt.Sprintf("completed_at = $%d", argIndex))
		args = append(args, req.CompletedAt)
		argIndex++
	}
	
	if req.AssigneeID != nil {
		setParts = append(setParts, fmt.Sprintf("assignee_id = $%d", argIndex))
		args = append(args, *req.AssigneeID)
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

// GetByProjectID retrieves all tasks for a specific project
func (r *TaskRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]domain.Task, error) {
	query := `
		SELECT id, title, description, status, priority, due_date, completed_at, assignee_id, project_id, milestone_id, created_at, updated_at
		FROM tasks
		WHERE project_id = $1
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanTasks(rows)
}

// GetByUserID retrieves all tasks assigned to a specific user
func (r *TaskRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Task, error) {
	query := `
		SELECT id, title, description, status, priority, due_date, completed_at, assignee_id, project_id, milestone_id, created_at, updated_at
		FROM tasks
		WHERE assignee_id = $1
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanTasks(rows)
}

// GetByMilestoneID retrieves all tasks for a specific milestone
func (r *TaskRepository) GetByMilestoneID(ctx context.Context, milestoneID uuid.UUID) ([]domain.Task, error) {
	query := `
		SELECT id, title, description, status, priority, due_date, completed_at, assignee_id, project_id, milestone_id, created_at, updated_at
		FROM tasks
		WHERE milestone_id = $1
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, milestoneID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanTasks(rows)
}

// GetByStatus retrieves all tasks with a specific status
func (r *TaskRepository) GetByStatus(ctx context.Context, status domain.TaskStatus) ([]domain.Task, error) {
	query := `
		SELECT id, title, description, status, priority, due_date, completed_at, assignee_id, project_id, milestone_id, created_at, updated_at
		FROM tasks
		WHERE status = $1
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanTasks(rows)
}

// AssignTask assigns a task to a user with optional updates
func (r *TaskRepository) AssignTask(ctx context.Context, taskID uuid.UUID, req *domain.AssignTaskRequest) (*domain.Task, error) {
	setParts := []string{"assignee_id = $2"}
	args := []interface{}{taskID, req.UserID}
	argIndex := 3
	
	if req.DueDate != nil {
		setParts = append(setParts, fmt.Sprintf("due_date = $%d", argIndex))
		args = append(args, req.DueDate)
		argIndex++
	}
	
	if req.CompletedAt != nil {
		setParts = append(setParts, fmt.Sprintf("completed_at = $%d", argIndex))
		args = append(args, req.CompletedAt)
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
	
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	
	query := fmt.Sprintf(`
		UPDATE tasks 
		SET %s
		WHERE id = $1
	`, strings.Join(setParts, ", "))
	
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	
	return r.GetByID(ctx, taskID)
}

// UpdateStatus updates only the status of a task
func (r *TaskRepository) UpdateStatus(ctx context.Context, taskID uuid.UUID, status domain.TaskStatus) error {
	query := `
		UPDATE tasks 
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	
	result, err := r.db.ExecContext(ctx, query, status, time.Now(), taskID)
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

// scanTasks is a helper method to scan multiple task rows
func (r *TaskRepository) scanTasks(rows *sql.Rows) ([]domain.Task, error) {
	var tasks []domain.Task
	
	for rows.Next() {
		var task domain.Task
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&task.DueDate,
			&task.CompletedAt,
			&task.AssigneeID,
			&task.ProjectID,
			&task.MilestoneID,
			&task.CreatedAt,
			&task.UpdatedAt,
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
