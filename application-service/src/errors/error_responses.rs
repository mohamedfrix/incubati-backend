use axum::{
    http::StatusCode,
    response::{IntoResponse, Response},
    Json,
};
use serde_json::json;

use super::error_types::AppError;

impl IntoResponse for AppError {
    fn into_response(self) -> Response {
        let (status, error_message) = match &self {
            AppError::Database(err) => {
                tracing::error!("Database error: {}", err);
                (StatusCode::INTERNAL_SERVER_ERROR, "Database error".to_string())
            }
            AppError::Validation(err) => {
                tracing::warn!("Validation error: {}", err);
                (StatusCode::BAD_REQUEST, err.as_str().to_string())
            }
            AppError::NotFound(err) => {
                tracing::info!("Not found: {}", err);
                (StatusCode::NOT_FOUND, err.as_str().to_string())
            }
            AppError::Conflict(err) => {
                tracing::warn!("Conflict: {}", err);
                (StatusCode::CONFLICT, err.as_str().to_string())
            }
            AppError::BadRequest(err) => {
                tracing::warn!("Bad request: {}", err);
                (StatusCode::BAD_REQUEST, err.as_str().to_string())
            }
            AppError::Internal(err) => {
                tracing::error!("Internal error: {}", err);
                (StatusCode::INTERNAL_SERVER_ERROR, "Internal server error".to_string())
            }
            AppError::Unauthorized(err) => {
                tracing::warn!("Unauthorized: {}", err);
                (StatusCode::UNAUTHORIZED, err.as_str().to_string())
            }
            AppError::FileError(err) => {
                tracing::error!("File error: {}", err);
                (StatusCode::INTERNAL_SERVER_ERROR, err.as_str().to_string())
            }
            AppError::PermissionDenied(err) => {
                tracing::error!("Permission denied: {}", err);
                (StatusCode::FORBIDDEN, err.as_str().to_string())
            }
            AppError::InvalidRole { role } => {
                tracing::error!("Invalid role: {}", role);
                (StatusCode::BAD_REQUEST, format!("Invalid user role: {}", role))
            }
        };

        let body = Json(json!({
            "error": error_message,
            "status": status.as_u16()
        }));

        (status, body).into_response()
    }
}
