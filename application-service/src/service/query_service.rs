use std::sync::Arc;
use tracing::info;
use validator::Validate;

use crate::dto::*;
use crate::models::*;
use crate::repository::*;
use crate::errors::{AppError, AppResult};

/// Query service for read operations and complex queries
pub struct QueryService {
    repository: Arc<dyn ApplicationRepository>,
}

impl QueryService {
    /// Create a new query service
    pub fn new(repository: Arc<dyn ApplicationRepository>) -> Self {
        Self { repository }
    }

    /// Get applications with filters and pagination
    pub async fn get_applications(
        &self,
        filters: ApplicationFilters,
    ) -> AppResult<PaginatedApplicationResponse> {
        info!("Fetching applications with filters: {:?}", filters);
        
        // Validate filters
        filters.validate().map_err(|e| {
            tracing::warn!("Validation failed for application filters: {:?}", e);
            AppError::Validation(format!("Invalid filters: {}", e))
        })?;

        // Validate pagination parameters
        let page = filters.page.unwrap_or(1);
        let limit = filters.limit.unwrap_or(20);
        
        if page < 1 {
            return Err(AppError::BadRequest("Page must be greater than 0".to_string()));
        }
        
        if limit < 1 || limit > 100 {
            return Err(AppError::BadRequest("Limit must be between 1 and 100".to_string()));
        }

        let offset = (page - 1) * limit;

        info!("Requesting page {} with limit {} (offset: {})", page, limit, offset);

        // TODO: Enhance repository to support application_type and status filtering
        // For now, we only filter by user_id in the repository and filter the rest in memory
        let (mut applications, total_count) = self
            .repository
            .get_applications(
                filters.user_id.as_deref(),
                limit * 2, // Get more records to account for filtering
                offset,
            )
            .await?;

        // Apply additional filters in memory (temporary solution)
        if let Some(ref app_type) = filters.application_type {
            applications.retain(|app| &app.application_type == app_type);
        }
        
        if let Some(ref status) = filters.status {
            applications.retain(|app| &app.status == status);
        }

        // Truncate to requested limit after filtering
        applications.truncate(limit as usize);

        info!("Found {} applications after filtering (total: {})", applications.len(), total_count);

        // Convert to response DTOs
        let results: Vec<ApplicationResponse> = applications
            .into_iter()
            .map(|app| self.convert_to_response(app))
            .collect();

        // Calculate pagination URLs with current filters
        let mut base_params = Vec::new();
        if let Some(ref user_id) = filters.user_id {
            base_params.push(format!("user_id={}", user_id));
        }
        if let Some(ref app_type) = filters.application_type {
            base_params.push(format!("application_type={}", app_type.to_db_string()));
        }
        if let Some(ref status) = filters.status {
            base_params.push(format!("status={}", status.to_db_string()));
        }
        base_params.push(format!("limit={}", limit));

        let base_query = if base_params.is_empty() {
            String::new()
        } else {
            format!("?{}", base_params.join("&"))
        };

        let next = if (page * limit) < total_count as u32 {
            Some(format!("{}&page={}", base_query, page + 1))
        } else {
            None
        };

        let previous = if page > 1 {
            Some(format!("{}&page={}", base_query, page - 1))
        } else {
            None
        };

        let response = PaginatedApplicationResponse {
            count: total_count,
            next,
            previous,
            results,
        };

        Ok(response)
    }

    /// Convert application model to response DTO
    fn convert_to_response(&self, app: Application) -> ApplicationResponse {
        ApplicationResponse {
            id: app.id,
            application_type: app.application_type.clone(),
            type_display: app.application_type.to_string(),
            status: app.status.clone(),
            status_display: app.status.to_string(),
            submission_date: app.submission_date,
            feedback: app.feedback,
            user_id: app.user_id,
            created_at: app.created_at,
            updated_at: app.updated_at,
            documents: Vec::new(),
            type_specific_data: None,
        }
    }
}
