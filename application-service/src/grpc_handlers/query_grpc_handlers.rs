/// gRPC handlers for application query operations
/// 
/// This module implements query-related gRPC operations such as:
/// - Listing applications with filtering and pagination
/// - Advanced search capabilities
/// - Aggregated data queries
/// 
/// These handlers follow the same architectural pattern as the REST handlers,
/// validating requests and calling the query service layer.

use tonic::{Request, Response, Status};
use tracing::{info, warn, error};
use validator::Validate;

use super::{GrpcServiceState, map_app_error_to_status, extract_request_metadata};
use super::conversions::{proto, *};
use crate::dto::*;

/// gRPC service implementation for query operations
pub struct QueryGrpcService {
    state: GrpcServiceState,
}

impl QueryGrpcService {
    /// Create a new QueryGrpcService
    pub fn new(state: GrpcServiceState) -> Self {
        Self { state }
    }

    /// List applications with filtering and pagination (implementation method)
    /// 
    /// This method provides the actual implementation for listing applications.
    /// It's separate to allow calling from the main ApplicationService as well.
    pub async fn list_applications_impl(
        &self,
        request: Request<proto::ListApplicationsRequest>,
    ) -> Result<Response<proto::PaginatedApplicationResponse>, Status> {
        let (user_agent, request_id) = extract_request_metadata(&request);
        let grpc_request = request.into_inner();
        
        info!(
            request_id = %request_id,
            user_agent = %user_agent,
            "Received gRPC list applications request"
        );

        // Convert and validate filter parameters
        let filters = self.convert_list_request_to_filters(grpc_request)?;

        // Validate the filters using existing validation rules
        filters.validate().map_err(|e| {
            warn!(
                request_id = %request_id,
                error = %e,
                "Validation failed for list applications request"
            );
            Status::invalid_argument(format!("Invalid query parameters: {}", e))
        })?;

        // Call service layer
        match self.state.query_service.get_applications(filters).await {
            Ok(paginated_response) => {
                info!(
                    request_id = %request_id,
                    count = paginated_response.count,
                    results_len = paginated_response.results.len(),
                    "Successfully retrieved applications via gRPC"
                );
                
                let grpc_response = self.convert_paginated_response_to_grpc(paginated_response);
                Ok(Response::new(grpc_response))
            }
            Err(e) => {
                error!(
                    request_id = %request_id,
                    error = %e,
                    "Failed to retrieve applications via gRPC"
                );
                Err(map_app_error_to_status(e))
            }
        }
    }

    /// Convert gRPC ListApplicationsRequest to internal ApplicationFilters
    fn convert_list_request_to_filters(
        &self,
        request: proto::ListApplicationsRequest,
    ) -> Result<ApplicationFilters, Status> {
        // Convert application type filter if provided
        let application_type = if let Some(type_value) = request.application_type {
            Some(grpc_to_application_type(type_value)?)
        } else {
            None
        };

        // Convert status filter if provided
        let status = if let Some(status_value) = request.status {
            Some(grpc_to_application_status(status_value)?)
        } else {
            None
        };

        // Set default pagination values if not provided
        let page = request.page.map(|v| v as u32).unwrap_or(1);
        let limit = request.page_size.map(|v| v as u32).unwrap_or(20);

        // Validate pagination parameters
        if page < 1 {
            return Err(Status::invalid_argument("Page must be greater than 0"));
        }
        if limit < 1 || limit > 100 {
            return Err(Status::invalid_argument("Page size must be between 1 and 100"));
        }

        Ok(ApplicationFilters {
            user_id: request.user_id,
            application_type,
            status,
            page: Some(page),
            limit: Some(limit),
        })
    }

    /// Convert internal PaginatedApplicationResponse to gRPC response
    fn convert_paginated_response_to_grpc(
        &self,
        response: PaginatedApplicationResponse,
    ) -> proto::PaginatedApplicationResponse {
        proto::PaginatedApplicationResponse {
            results: response.results
                .iter()
                .map(application_response_to_grpc)
                .collect(),
            total_count: response.count as i32,
            page: 1, // Default since our internal structure doesn't track current page
            page_size: response.results.len() as i32,
            total_pages: 1, // Default since we don't calculate total pages
        }
    }
}

/// Additional query operations that could be added in the future
impl QueryGrpcService {
    /// Get application statistics (for future implementation)
    /// 
    /// This could provide aggregated statistics about applications such as:
    /// - Count by status
    /// - Count by type
    /// - Monthly submission trends
    /// - Average processing time
    #[allow(dead_code)]
    async fn get_application_statistics(
        &self,
        _request: Request<()>, // Define appropriate request type
    ) -> Result<Response<()>, Status> { // Define appropriate response type
        // Implementation placeholder for future enhancement
        Err(Status::unimplemented("Statistics endpoint not yet implemented"))
    }

    /// Search applications by text (for future implementation)
    /// 
    /// This could provide full-text search capabilities across:
    /// - Application descriptions
    /// - Company names
    /// - Project titles
    /// - User feedback
    #[allow(dead_code)]
    async fn search_applications(
        &self,
        _request: Request<()>, // Define appropriate search request type
    ) -> Result<Response<proto::PaginatedApplicationResponse>, Status> {
        // Implementation placeholder for future enhancement
        Err(Status::unimplemented("Search endpoint not yet implemented"))
    }

    /// Get application timeline (for future implementation)
    /// 
    /// This could provide a detailed timeline of status changes and events
    /// for a specific application.
    #[allow(dead_code)]
    async fn get_application_timeline(
        &self,
        _request: Request<proto::GetApplicationRequest>,
    ) -> Result<Response<()>, Status> { // Define appropriate timeline response type
        // Implementation placeholder for future enhancement
        Err(Status::unimplemented("Timeline endpoint not yet implemented"))
    }
}
