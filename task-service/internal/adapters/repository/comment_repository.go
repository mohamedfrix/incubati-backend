package repository

import (
	"context"
	"database/sql"
	"math"
	"time"

	"moulaybdl/zindy/task_service/internal/core/domain"
	"moulaybdl/zindy/task_service/internal/core/ports"

	"github.com/google/uuid"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) ports.CommentRepository {
	return &CommentRepository{db: db}
}

// Create creates a new task comment and returns the created comment
func (r *CommentRepository) Create(ctx context.Context, taskID uuid.UUID, comment *domain.TaskComment) (*domain.TaskComment, error) {
	query := `
		INSERT INTO task_comments (id, task_id, author_user_id, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, task_id, author_user_id, content, created_at, updated_at
	`
	
	now := time.Now()
	comment.ID = uuid.New()
	comment.TaskID = taskID
	comment.CreatedAt = now
	comment.UpdatedAt = now
	
	row := r.db.QueryRowContext(ctx, query,
		comment.ID,
		comment.TaskID,
		comment.AuthorUserID,
		comment.Content,
		comment.CreatedAt,
		comment.UpdatedAt,
	)
	
	var createdComment domain.TaskComment
	err := row.Scan(
		&createdComment.ID,
		&createdComment.TaskID,
		&createdComment.AuthorUserID,
		&createdComment.Content,
		&createdComment.CreatedAt,
		&createdComment.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &createdComment, nil
}

// GetByID retrieves a comment by its ID
func (r *CommentRepository) GetByID(ctx context.Context, commentID uuid.UUID) (*domain.TaskComment, error) {
	query := `
		SELECT id, task_id, author_user_id, content, created_at, updated_at
		FROM task_comments
		WHERE id = $1
	`
	
	row := r.db.QueryRowContext(ctx, query, commentID)
	
	var comment domain.TaskComment
	err := row.Scan(
		&comment.ID,
		&comment.TaskID,
		&comment.AuthorUserID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrCommentNotFound
		}
		return nil, err
	}
	
	return &comment, nil
}

// Update updates a comment's content
func (r *CommentRepository) Update(ctx context.Context, commentID uuid.UUID, req *domain.UpdateCommentRequest) (*domain.TaskComment, error) {
	query := `
		UPDATE task_comments 
		SET content = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, task_id, author_user_id, content, created_at, updated_at
	`
	
	now := time.Now()
	
	row := r.db.QueryRowContext(ctx, query, req.Content, now, commentID)
	
	var comment domain.TaskComment
	err := row.Scan(
		&comment.ID,
		&comment.TaskID,
		&comment.AuthorUserID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrCommentNotFound
		}
		return nil, err
	}
	
	return &comment, nil
}

// Delete deletes a comment
func (r *CommentRepository) Delete(ctx context.Context, commentID uuid.UUID) error {
	query := `DELETE FROM task_comments WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, commentID)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return domain.ErrCommentNotFound
	}
	
	return nil
}

// GetByTaskID retrieves all comments for a task with pagination
func (r *CommentRepository) GetByTaskID(ctx context.Context, taskID uuid.UUID, pagination *domain.PaginationInfo) ([]domain.TaskComment, *domain.PaginationInfo, error) {
	// Set default pagination if not provided
	if pagination == nil {
		pagination = &domain.PaginationInfo{
			Page:     1,
			PageSize: 20,
		}
	}
	
	// Validate pagination
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.PageSize < 1 || pagination.PageSize > 100 {
		pagination.PageSize = 20
	}
	
	// First, get total count
	countQuery := `SELECT COUNT(*) FROM task_comments WHERE task_id = $1`
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, taskID).Scan(&totalCount)
	if err != nil {
		return nil, nil, err
	}
	
	// Calculate pagination
	offset := (pagination.Page - 1) * pagination.PageSize
	totalPages := int(math.Ceil(float64(totalCount) / float64(pagination.PageSize)))
	
	// Get comments with pagination
	query := `
		SELECT id, task_id, author_user_id, content, created_at, updated_at
		FROM task_comments
		WHERE task_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.QueryContext(ctx, query, taskID, pagination.PageSize, offset)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	
	var comments []domain.TaskComment
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
			return nil, nil, err
		}
		comments = append(comments, comment)
	}
	
	if err = rows.Err(); err != nil {
		return nil, nil, err
	}
	
	// Update pagination info
	paginationInfo := &domain.PaginationInfo{
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}
	
	return comments, paginationInfo, nil
}

// GetCommentCount returns the number of comments for a task
func (r *CommentRepository) GetCommentCount(ctx context.Context, taskID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM task_comments WHERE task_id = $1`
	
	var count int
	err := r.db.QueryRowContext(ctx, query, taskID).Scan(&count)
	if err != nil {
		return 0, err
	}
	
	return count, nil
}

// ExistsByID checks if a comment exists by ID
func (r *CommentRepository) ExistsByID(ctx context.Context, commentID uuid.UUID) (bool, error) {
	query := `SELECT 1 FROM task_comments WHERE id = $1 LIMIT 1`
	
	var exists int
	err := r.db.QueryRowContext(ctx, query, commentID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	
	return exists == 1, nil
}

// IsAuthor checks if a user is the author of a comment
func (r *CommentRepository) IsAuthor(ctx context.Context, commentID uuid.UUID, userID uuid.UUID) (bool, error) {
	query := `SELECT 1 FROM task_comments WHERE id = $1 AND author_user_id = $2 LIMIT 1`
	
	var exists int
	err := r.db.QueryRowContext(ctx, query, commentID, userID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	
	return exists == 1, nil
}
