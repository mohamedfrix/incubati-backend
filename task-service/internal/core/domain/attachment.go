package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TaskAttachment represents a file attachment on a task
type TaskAttachment struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	TaskID             uuid.UUID `json:"task_id" db:"task_id"`
	UploadedByUserID   uuid.UUID `json:"uploaded_by_user_id" db:"uploaded_by_user_id"`
	FileName           string    `json:"file_name" db:"file_name"`
	FileURL            string    `json:"file_url" db:"file_url"`
	FileSize           int64     `json:"file_size" db:"file_size"`
	MimeType           string    `json:"mime_type" db:"mime_type"`
	UploadedAt         time.Time `json:"uploaded_at" db:"uploaded_at"`
}

// CreateAttachmentRequest represents the request payload for creating an attachment
type CreateAttachmentRequest struct {
	UploadedByUserID *uuid.UUID `json:"uploaded_by_user_id"` // Auto-set if not provided
	FileName         string     `json:"file_name" validate:"required,max=255"`
	FileURL          string     `json:"file_url" validate:"required,url"`
	FileSize         int64      `json:"file_size" validate:"required,gt=0,lte=52428800"` // Max 50MB
	MimeType         string     `json:"mime_type" validate:"required,max=100"`
}

// AttachmentResponse represents the response structure for attachment operations
type AttachmentResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    *TaskAttachment `json:"data,omitempty"`
}

// AttachmentListResponse represents the response structure for attachment lists
type AttachmentListResponse struct {
	Success bool               `json:"success"`
	Data    AttachmentListData `json:"data"`
}

// AttachmentListData represents the data structure for attachment list responses
type AttachmentListData struct {
	Attachments []TaskAttachment `json:"attachments"`
	Pagination  PaginationInfo   `json:"pagination"`
}

// AttachmentWithDisplayData represents an attachment with additional display information
type AttachmentWithDisplayData struct {
	TaskAttachment
	FileSizeDisplay string `json:"file_size_display"`
}

// AllowedMimeTypes defines the allowed file types for attachments
var AllowedMimeTypes = map[string]bool{
	// Images
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
	
	// Documents
	"application/pdf":   true,
	"text/plain":        true,
	"text/csv":          true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
	
	// Archives
	"application/zip": true,
}

// MaxFileSize defines the maximum allowed file size (50MB)
const MaxFileSize = 52428800

// IsValidMimeType checks if the mime type is allowed
func IsValidMimeType(mimeType string) bool {
	return AllowedMimeTypes[strings.ToLower(mimeType)]
}

// IsValidFileSize checks if the file size is within limits
func IsValidFileSize(size int64) bool {
	return size > 0 && size <= MaxFileSize
}

// FormatFileSize formats file size in human readable format
func FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// GetDisplayData returns attachment with additional display information
func (a *TaskAttachment) GetDisplayData() AttachmentWithDisplayData {
	return AttachmentWithDisplayData{
		TaskAttachment:  *a,
		FileSizeDisplay: FormatFileSize(a.FileSize),
	}
}

// Validate validates the attachment data
func (req *CreateAttachmentRequest) Validate() error {
	if !IsValidMimeType(req.MimeType) {
		return fmt.Errorf("mime type %s is not allowed", req.MimeType)
	}
	
	if !IsValidFileSize(req.FileSize) {
		return fmt.Errorf("file size %d exceeds maximum allowed size of %d bytes", req.FileSize, MaxFileSize)
	}
	
	return nil
}
