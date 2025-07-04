use std::sync::Arc;
use tracing::{info, warn};
use uuid::Uuid;
use bytes::Bytes;

use crate::{
    dto::responses::{DocumentResponse, DocumentDownloadResponse},
    errors::{AppError, AppResult},
    models::Document,
    repository::{ApplicationRepository, DocumentRepository},
    utils::MinioClient,
};
use super::permissions::{PermissionValidator, DocumentPermission};

/// Service for handling document operations
pub struct DocumentService {
    repository: Arc<dyn ApplicationRepository>,
    document_repository: Arc<dyn DocumentRepository>,
    minio_client: Arc<MinioClient>,
    permission_validator: PermissionValidator,
}

impl DocumentService {
    /// Create a new DocumentService
    pub fn new(
        repository: Arc<dyn ApplicationRepository>,
        document_repository: Arc<dyn DocumentRepository>,
        minio_client: Arc<MinioClient>,
    ) -> Self {
        let permission_validator = PermissionValidator::new(
            repository.clone(),
            document_repository.clone(),
        );
        
        Self { 
            repository, 
            document_repository,
            minio_client,
            permission_validator,
        }
    }

    /// Upload a document file to Minio and create database record (with permission validation)
    pub async fn upload_document_with_permission(
        &self,
        application_id: &str,
        original_name: String,
        file_content: Bytes,
        content_type: String,
        document_type: String,
        requesting_user_id: &str,
        requesting_user_role: &str,
    ) -> AppResult<DocumentResponse> {
        info!("Uploading document with permission validation for application {}", application_id);
        
        // Validate permissions first
        let _app = self.permission_validator
            .validate_document_creation_permission(
                application_id,
                requesting_user_id,
                requesting_user_role,
            )
            .await?;
        
        // Parse UUID
        let app_uuid = Uuid::parse_str(application_id)
            .map_err(|_| AppError::BadRequest("Invalid application ID format".to_string()))?;
        
        // Call the existing upload method
        self.upload_document(app_uuid, original_name, file_content, content_type, document_type).await
    }

    /// Upload a document file to Minio and create database record
    pub async fn upload_document(
        &self,
        application_id: Uuid,
        original_name: String,
        file_content: Bytes,
        content_type: String,
        document_type: String,
    ) -> AppResult<DocumentResponse> {
        info!("Uploading document for application {}", application_id);

        // Verify application exists
        let application = self
            .repository
            .get_application_by_id(application_id)
            .await?
            .ok_or_else(|| {
                warn!("Application {} not found for document upload", application_id);
                AppError::NotFound(format!("Application with id {} not found", application_id))
            })?;

        info!("Found application {} for document upload", application.id);

        let document_id = Uuid::new_v4();
        
        // Generate filename for Minio storage
        let file_extension = original_name.split('.').last().unwrap_or("bin");
        let filename = format!("applications/{}/documents/{}.{}", 
            application_id, document_id, file_extension);
        
        // Upload file to Minio
        let _minio_path = self
            .minio_client
            .upload_document(
                application_id,
                document_id,
                file_content.clone(),
                &content_type,
                file_extension,
            )
            .await?;

        // Get file size
        let file_size = file_content.len() as i64;

        // Create document record
        let document = Document {
            id: document_id,
            application_id,
            filename: filename.clone(),
            original_name: original_name.clone(),
            content_type: content_type.clone(),
            file_size,
            document_type: document_type.clone(),
            uploaded_at: chrono::Utc::now(),
        };

        // Create document in repository
        let created_document = self.document_repository.create_document(&document).await?;
        
        info!("Document {} uploaded successfully for application {}", created_document.id, application_id);

        // Generate download URL for the response
        let download_url = self
            .minio_client
            .generate_document_download_url(&created_document.filename, None)
            .await
            .unwrap_or_else(|_| format!("/api/documents/{}/download", created_document.id));

        Ok(DocumentResponse {
            id: created_document.id,
            application_id: created_document.application_id,
            filename: created_document.filename,
            original_name: created_document.original_name,
            content_type: created_document.content_type,
            file_size: created_document.file_size,
            document_type: created_document.document_type,
            uploaded_at: created_document.uploaded_at,
            download_url,
        })
    }

    /// Get documents for an application
    pub async fn get_application_documents(
        &self,
        application_id: Uuid,
    ) -> AppResult<Vec<DocumentResponse>> {
        info!("Fetching documents for application {}", application_id);

        // Verify application exists
        let _application = self
            .repository
            .get_application_by_id(application_id)
            .await?
            .ok_or_else(|| {
                warn!("Application {} not found", application_id);
                AppError::NotFound(format!("Application with id {} not found", application_id))
            })?;

        // Get documents from repository
        let documents = self.document_repository.get_documents_by_application_id(application_id).await?;
        
        info!("Found {} documents for application {}", documents.len(), application_id);
        
        let mut document_responses = Vec::new();
        
        for document in documents {
            // Generate download URL for each document
            let download_url = self
                .minio_client
                .generate_document_download_url(&document.filename, None)
                .await
                .unwrap_or_else(|_| format!("/api/documents/{}/download", document.id));

            document_responses.push(DocumentResponse {
                id: document.id,
                application_id: document.application_id,
                filename: document.filename,
                original_name: document.original_name,
                content_type: document.content_type,
                file_size: document.file_size,
                document_type: document.document_type,
                uploaded_at: document.uploaded_at,
                download_url,
            });
        }
            
        Ok(document_responses)
    }

    /// Get a specific document (with permission validation)
    pub async fn get_document_with_permission(
        &self,
        document_id: &str,
        requesting_user_id: &str,
        requesting_user_role: &str,
    ) -> AppResult<DocumentResponse> {
        info!("Getting document with permission validation: {}", document_id);
        
        // Validate permissions first
        let _ = self.permission_validator
            .validate_document_permission(
                document_id,
                requesting_user_id,
                requesting_user_role,
                DocumentPermission::Read,
            )
            .await?;
        
        // Parse UUID
        let doc_uuid = Uuid::parse_str(document_id)
            .map_err(|_| AppError::BadRequest("Invalid document ID format".to_string()))?;
        
        // Call the existing get method
        self.get_document(doc_uuid).await
    }

    /// Get a specific document
    pub async fn get_document(&self, document_id: Uuid) -> AppResult<DocumentResponse> {
        info!("Fetching document {}", document_id);

        // Get document from repository
        let document = self.document_repository.get_document_by_id(document_id).await?
            .ok_or_else(|| AppError::NotFound(format!("Document with id {} not found", document_id)))?;

        // Generate download URL
        let download_url = self
            .minio_client
            .generate_document_download_url(&document.filename, None)
            .await
            .unwrap_or_else(|_| format!("/api/documents/{}/download", document.id));

        Ok(DocumentResponse {
            id: document.id,
            application_id: document.application_id,
            filename: document.filename,
            original_name: document.original_name,
            content_type: document.content_type,
            file_size: document.file_size,
            document_type: document.document_type,
            uploaded_at: document.uploaded_at,
            download_url,
        })
    }

    /// Generate a download URL for a document (with permission validation)
    pub async fn generate_download_url_with_permission(
        &self,
        document_id: &str,
        requesting_user_id: &str,
        requesting_user_role: &str,
        expiry_seconds: Option<u64>,
    ) -> AppResult<DocumentDownloadResponse> {
        info!("Generating download URL with permission validation for document: {}", document_id);
        
        // Validate permissions first
        let _ = self.permission_validator
            .validate_document_permission(
                document_id,
                requesting_user_id,
                requesting_user_role,
                DocumentPermission::Read,
            )
            .await?;
        
        // Parse UUID
        let doc_uuid = Uuid::parse_str(document_id)
            .map_err(|_| AppError::BadRequest("Invalid document ID format".to_string()))?;
        
        // Call the existing generate_download_url method
        self.generate_download_url(doc_uuid, expiry_seconds).await
    }

    /// Generate a download URL for a document
    pub async fn generate_download_url(
        &self, 
        document_id: Uuid,
        expiry_seconds: Option<u64>
    ) -> AppResult<DocumentDownloadResponse> {
        info!("Generating download URL for document {}", document_id);

        // Get document from repository
        let document = self.document_repository.get_document_by_id(document_id).await?
            .ok_or_else(|| AppError::NotFound(format!("Document with id {} not found", document_id)))?;

        // Generate download URL from Minio
        let download_url = self
            .minio_client
            .generate_document_download_url(&document.filename, expiry_seconds)
            .await?;

        Ok(DocumentDownloadResponse { download_url })
    }

    /// Download a document (legacy method for compatibility)
    pub async fn download_document(&self, document_id: Uuid) -> AppResult<DocumentDownloadResponse> {
        self.generate_download_url(document_id, None).await
    }

    /// Download document content directly
    pub async fn download_document_content(&self, document_id: Uuid) -> AppResult<(Bytes, String, String)> {
        info!("Downloading document content {}", document_id);

        // Get document from repository
        let document = self.document_repository.get_document_by_id(document_id).await?
            .ok_or_else(|| AppError::NotFound(format!("Document with id {} not found", document_id)))?;

        // Download file content from Minio
        let content = self
            .minio_client
            .download_file(self.minio_client.documents_bucket(), &document.filename)
            .await?;

        Ok((content, document.content_type, document.original_name))
    }

    /// Delete a document (with permission validation)
    pub async fn delete_document_with_permission(
        &self,
        document_id: &str,
        requesting_user_id: &str,
        requesting_user_role: &str,
    ) -> AppResult<()> {
        info!("Deleting document with permission validation: {}", document_id);
        
        // Validate permissions first
        let _ = self.permission_validator
            .validate_document_permission(
                document_id,
                requesting_user_id,
                requesting_user_role,
                DocumentPermission::Delete,
            )
            .await?;
        
        // Parse UUID
        let doc_uuid = Uuid::parse_str(document_id)
            .map_err(|_| AppError::BadRequest("Invalid document ID format".to_string()))?;
        
        // Call the existing delete method
        self.delete_document(doc_uuid).await
    }

    /// Delete a document
    pub async fn delete_document(&self, document_id: Uuid) -> AppResult<()> {
        info!("Deleting document {}", document_id);

        // Get document from repository first
        let document = self.document_repository.get_document_by_id(document_id).await?
            .ok_or_else(|| AppError::NotFound(format!("Document with id {} not found", document_id)))?;

        // Delete file from Minio
        if let Err(e) = self.minio_client.delete_document(&document.filename).await {
            warn!("Failed to delete file from Minio: {}, continuing with database deletion", e);
        }

        // Delete document from repository
        let deleted = self.document_repository.delete_document(document_id).await?;
        
        if deleted {
            info!("Successfully deleted document {}", document_id);
            Ok(())
        } else {
            Err(AppError::NotFound(format!("Document with id {} not found", document_id)))
        }
    }
}
