use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::Json,
};
use serde_json::{json, Value};
use uuid::Uuid;
use tracing::{info, warn};
use validator::Validate;

use crate::dto::*;
use crate::errors::AppError;
use super::AppState;

/// Create a new application
pub async fn create_application(
    State(state): State<AppState>,
    Json(request): Json<CreateApplicationRequest>,
) -> Result<(StatusCode, Json<ApplicationResponse>), AppError> {
    info!("Received request to create application for user: {}", request.user_id);

    // Validate request data
    request.validate().map_err(|e| {
        warn!("Invalid create application request: {:?}", e);
        AppError::Validation(format!("Invalid request data: {}", e))
    })?;

    let result = state.application_service.create_application(request).await?;
    
    info!("Successfully created application with ID: {}", result.id);
    Ok((StatusCode::CREATED, Json(result)))
}

/// Get a specific application by ID
pub async fn get_application(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<Json<ApplicationResponse>, AppError> {
    info!("Received request to get application: {}", id);

    match state.application_service.get_application_by_id(id).await? {
        Some(application) => {
            info!("Successfully retrieved application: {}", id);
            Ok(Json(application))
        }
        None => {
            warn!("Application not found: {}", id);
            Err(AppError::NotFound(format!("Application with ID {} not found", id)))
        }
    }
}

/// Update an application (full update)
pub async fn update_application(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(request): Json<UpdateApplicationRequest>,
) -> Result<Json<ApplicationResponse>, AppError> {
    info!("Received request to update application: {}", id);

    // Validate request data
    request.validate().map_err(|e| {
        warn!("Invalid update application request: {:?}", e);
        AppError::Validation(format!("Invalid request data: {}", e))
    })?;

    match state.application_service.update_application(id, request).await? {
        Some(application) => {
            info!("Successfully updated application: {}", id);
            Ok(Json(application))
        }
        None => {
            warn!("Application not found for update: {}", id);
            Err(AppError::NotFound(format!("Application with ID {} not found", id)))
        }
    }
}

/// Change application status (partial update)
pub async fn change_application_status(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(request): Json<ChangeStatusRequest>,
) -> Result<Json<Value>, AppError> {
    info!("Received request to change status for application: {}", id);

    // Validate request data
    request.validate().map_err(|e| {
        warn!("Invalid status change request: {:?}", e);
        AppError::Validation(format!("Invalid request data: {}", e))
    })?;

    match state.application_service.update_application_status(id, request).await? {
        Some(application) => {
            info!("Successfully updated status for application: {}", id);
            Ok(Json(json!({
                "status": application.status,
                "status_display": application.status_display,
                "feedback": application.feedback
            })))
        }
        None => {
            warn!("Application not found for status update: {}", id);
            Err(AppError::NotFound(format!("Application with ID {} not found", id)))
        }
    }
}

/// Delete an application
pub async fn delete_application(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<StatusCode, AppError> {
    info!("Received request to delete application: {}", id);

    let deleted = state.application_service.delete_application(id).await?;
    
    if deleted {
        info!("Successfully deleted application: {}", id);
        Ok(StatusCode::NO_CONTENT)
    } else {
        warn!("Application not found for deletion: {}", id);
        Err(AppError::NotFound(format!("Application with ID {} not found", id)))
    }
}
