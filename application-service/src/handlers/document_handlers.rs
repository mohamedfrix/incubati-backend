use axum::{
    extract::{Path, State, Multipart},
    http::{StatusCode, HeaderMap, header},
    response::{IntoResponse, Response},
    Json,
};
use bytes::Bytes;
use tracing::{info, warn};
use uuid::Uuid;

use crate::{
    dto::responses::{DocumentResponse, DocumentDownloadResponse},
    errors::AppError,
};

use super::AppState;

/// Get documents for an application
pub async fn get_application_documents(
    State(state): State<AppState>,
    Path(application_id): Path<Uuid>,
) -> Result<Json<Vec<DocumentResponse>>, AppError> {
    info!("GET /applications/{}/documents", application_id);
    
    let documents = state
        .document_service
        .get_application_documents(application_id)
        .await?;
    
    Ok(Json(documents))
}

/// Upload a document for an application
pub async fn upload_document(
    State(state): State<AppState>,
    Path(application_id): Path<Uuid>,
    mut multipart: Multipart,
) -> Result<Json<DocumentResponse>, AppError> {
    info!("POST /applications/{}/documents", application_id);
    
    let mut original_name: Option<String> = None;
    let mut file_content: Option<Bytes> = None;
    let mut content_type: Option<String> = None;
    let mut document_type: Option<String> = None;
    
    // Process multipart form data
    while let Some(field) = multipart.next_field().await
        .map_err(|e| AppError::Validation(format!("Invalid multipart data: {}", e)))? 
    {
        let field_name = field.name().unwrap_or("");
        
        match field_name {
            "original_name" => {
                let field_data = field.text().await
                    .map_err(|e| AppError::Validation(format!("Invalid original_name field: {}", e)))?;
                original_name = Some(field_data);
            }
            "document_type" => {
                let field_data = field.text().await
                    .map_err(|e| AppError::Validation(format!("Invalid document_type field: {}", e)))?;
                document_type = Some(field_data);
            }
            "file" => {
                // Get file metadata
                if let Some(file_name) = field.file_name() {
                    // Use original filename if original_name not provided
                    if original_name.is_none() {
                        original_name = Some(file_name.to_string());
                    }
                }
                
                content_type = field.content_type().map(|ct| ct.to_string());
                
                // Read file content
                let data = field.bytes().await
                    .map_err(|e| AppError::Validation(format!("Failed to read file: {}", e)))?;
                file_content = Some(data);
            }
            _ => {
                // Skip unknown fields
                warn!("Unknown multipart field: {}", field_name);
            }
        }
    }
    
    // Validate required fields
    let original_name = original_name.ok_or_else(|| AppError::Validation("Original name is required".to_string()))?;
    let file_content = file_content.ok_or_else(|| AppError::Validation("File is required".to_string()))?;
    let content_type = content_type.unwrap_or_else(|| "application/octet-stream".to_string());
    let document_type = document_type.unwrap_or_else(|| "general".to_string());
    
    // Upload document
    let document = state
        .document_service
        .upload_document(application_id, original_name, file_content, content_type, document_type)
        .await?;
    
    Ok(Json(document))
}

/// Get a specific document
pub async fn get_document(
    State(state): State<AppState>,
    Path(document_id): Path<Uuid>,
) -> Result<Json<DocumentResponse>, AppError> {
    info!("GET /documents/{}", document_id);
    
    let document = state
        .document_service
        .get_document(document_id)
        .await?;
    
    Ok(Json(document))
}

/// Generate download URL for a document
pub async fn download_document(
    State(state): State<AppState>,
    Path(document_id): Path<Uuid>,
) -> Result<Json<DocumentDownloadResponse>, AppError> {
    info!("GET /documents/{}/download", document_id);
    
    let download_response = state
        .document_service
        .generate_download_url(document_id, None)
        .await?;
    
    Ok(Json(download_response))
}

/// Download document content directly (alternative endpoint)
pub async fn download_document_content(
    State(state): State<AppState>,
    Path(document_id): Path<Uuid>,
) -> Result<Response, AppError> {
    info!("GET /documents/{}/content", document_id);
    
    let (content, content_type, original_name) = state
        .document_service
        .download_document_content(document_id)
        .await?;
    
    let mut headers = HeaderMap::new();
    headers.insert(
        header::CONTENT_TYPE,
        content_type.parse().unwrap_or_else(|_| "application/octet-stream".parse().unwrap()),
    );
    headers.insert(
        header::CONTENT_DISPOSITION,
        format!("attachment; filename=\"{}\"", original_name)
            .parse()
            .unwrap_or_else(|_| "attachment".parse().unwrap()),
    );
    
    Ok((headers, content).into_response())
}

/// Delete a document
pub async fn delete_document(
    State(state): State<AppState>,
    Path(document_id): Path<Uuid>,
) -> Result<StatusCode, AppError> {
    info!("DELETE /documents/{}", document_id);
    
    state
        .document_service
        .delete_document(document_id)
        .await?;
    
    Ok(StatusCode::NO_CONTENT)
}
