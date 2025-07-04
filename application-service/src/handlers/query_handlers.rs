use axum::{
    extract::{Query, State},
    response::Json,
};
use tracing::{info, warn};
use validator::Validate;

use crate::dto::*;
use crate::errors::AppError;
use super::AppState;

/// List applications with optional filtering and pagination
pub async fn list_applications(
    State(state): State<AppState>,
    Query(filters): Query<ApplicationFilters>,
) -> Result<Json<PaginatedApplicationResponse>, AppError> {
    info!("Received request to list applications");

    // Validate query parameters
    filters.validate().map_err(|e| {
        warn!("Invalid query parameters: {:?}", e);
        AppError::Validation(format!("Invalid query parameters: {}", e))
    })?;

    let result = state.query_service.get_applications(filters).await?;
    
    info!("Successfully retrieved {} applications", result.results.len());
    Ok(Json(result))
}
