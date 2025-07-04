/// gRPC handlers for application management operations
/// 
/// This module implements the ApplicationService gRPC service, handling:
/// - Creating new applications
/// - Retrieving applications by ID
/// - Updating applications
/// - Changing application status
/// - Deleting applications
/// 
/// Each handler follows the same pattern:
/// 1. Extract and validate request metadata
/// 2. Validate and convert gRPC request to internal DTOs
/// 3. Call the appropriate service layer method
/// 4. Convert response back to gRPC types
/// 5. Handle errors with appropriate gRPC status codes

use tonic::{Request, Response, Status};
use tracing::{info, warn, error};
use validator::Validate;

use super::{GrpcServiceState, map_app_error_to_status, extract_request_metadata};
use super::conversions::{proto, *};
use crate::dto::*;

/// gRPC service implementation for application management
pub struct ApplicationGrpcService {
    state: GrpcServiceState,
}

impl ApplicationGrpcService {
    /// Create a new ApplicationGrpcService
    pub fn new(state: GrpcServiceState) -> Self {
        Self { state }
    }
}

#[tonic::async_trait]
impl proto::application_service_server::ApplicationService for ApplicationGrpcService {
    /// Create a new application
    /// 
    /// Validates the request data using the same validation rules as the REST API,
    /// then calls the service layer to create the application.
    async fn create_application(
        &self,
        request: Request<proto::CreateApplicationRequest>,
    ) -> Result<Response<proto::ApplicationResponse>, Status> {
        let (user_agent, request_id) = extract_request_metadata(&request);
        let grpc_request = request.into_inner();
        
        info!(
            request_id = %request_id,
            user_agent = %user_agent,
            user_id = %grpc_request.user_id,
            "Received gRPC create application request"
        );

        // Validate and convert gRPC request to internal DTO
        let application_type = grpc_to_application_type(grpc_request.application_type)?;
        let status = grpc_to_application_status(grpc_request.status)?;
        let type_specific_data = grpc_to_type_specific_data(
            grpc_request.type_specific_data, 
            &application_type
        )?;

        // Create internal request DTO
        let create_request = CreateApplicationRequest {
            user_id: grpc_request.user_id.clone(),
            application_type,
            status,
            type_specific_data,
        };

        // Validate the request using existing validation rules
        create_request.validate().map_err(|e| {
            warn!(
                request_id = %request_id,
                error = %e,
                "Validation failed for create application request"
            );
            Status::invalid_argument(format!("Invalid request data: {}", e))
        })?;

        // Call service layer
        match self.state.application_service.create_application(create_request).await {
            Ok(application_response) => {
                info!(
                    request_id = %request_id,
                    application_id = %application_response.id,
                    "Successfully created application via gRPC"
                );
                
                let grpc_response = application_response_to_grpc(&application_response);
                Ok(Response::new(grpc_response))
            }
            Err(e) => {
                error!(
                    request_id = %request_id,
                    error = %e,
                    "Failed to create application via gRPC"
                );
                Err(map_app_error_to_status(e))
            }
        }
    }

    /// Get an application by ID
    /// 
    /// Retrieves a specific application by its UUID, returning detailed information
    /// including type-specific data and associated documents.
    async fn get_application(
        &self,
        request: Request<proto::GetApplicationRequest>,
    ) -> Result<Response<proto::ApplicationResponse>, Status> {
        let (user_agent, request_id) = extract_request_metadata(&request);
        let grpc_request = request.into_inner();
        
        info!(
            request_id = %request_id,
            user_agent = %user_agent,
            application_id = %grpc_request.id,
            "Received gRPC get application request"
        );

        // Parse and validate UUID
        let id = parse_uuid(&grpc_request.id)?;

        // Call service layer
        match self.state.application_service.get_application_by_id(id).await {
            Ok(Some(application_response)) => {
                info!(
                    request_id = %request_id,
                    application_id = %id,
                    "Successfully retrieved application via gRPC"
                );
                
                let grpc_response = application_response_to_grpc(&application_response);
                Ok(Response::new(grpc_response))
            }
            Ok(None) => {
                warn!(
                    request_id = %request_id,
                    application_id = %id,
                    "Application not found via gRPC"
                );
                Err(Status::not_found(format!("Application with ID {} not found", id)))
            }
            Err(e) => {
                error!(
                    request_id = %request_id,
                    application_id = %id,
                    error = %e,
                    "Failed to retrieve application via gRPC"
                );
                Err(map_app_error_to_status(e))
            }
        }
    }

    /// Update an application (full update)
    /// 
    /// Updates an existing application with new data. This performs a full update
    /// of the application's status, feedback, and type-specific data.
    /// Now includes permission validation to prevent security vulnerabilities.
    async fn update_application(
        &self,
        request: Request<proto::UpdateApplicationRequest>,
    ) -> Result<Response<proto::ApplicationResponse>, Status> {
        let (user_agent, request_id) = extract_request_metadata(&request);
        let grpc_request = request.into_inner();
        
        info!(
            request_id = %request_id,
            user_agent = %user_agent,
            application_id = %grpc_request.id,
            requesting_user_id = %grpc_request.requesting_user_id,
            requesting_user_role = %grpc_request.requesting_user_role,
            "Received gRPC update application request"
        );

        // Validate required permission fields
        if grpc_request.requesting_user_id.is_empty() || grpc_request.requesting_user_role.is_empty() {
            return Err(Status::invalid_argument("requesting_user_id and requesting_user_role are required"));
        }

        // Convert status if provided
        let status = if let Some(status_value) = grpc_request.status {
            Some(grpc_to_application_status(status_value)?)
        } else {
            None
        };

        // Convert type-specific data if provided
        let type_specific_data = if let Some(data) = grpc_request.type_specific_data {
            // We need to get the current application to know its type
            let id = parse_uuid(&grpc_request.id)?;
            match self.state.application_service.get_application_by_id(id).await {
                Ok(Some(current_app)) => {
                    Some(grpc_to_type_specific_data(Some(data), &current_app.application_type)?)
                }
                Ok(None) => {
                    return Err(Status::not_found(format!("Application with ID {} not found", id)));
                }
                Err(e) => {
                    return Err(map_app_error_to_status(e));
                }
            }
        } else {
            None
        };

        // Create internal request DTO
        let update_request = UpdateApplicationRequest {
            status,
            feedback: grpc_request.feedback,
            type_specific_data,
        };

        // Validate the request using existing validation rules
        update_request.validate().map_err(|e| {
            warn!(
                request_id = %request_id,
                error = %e,
                "Validation failed for update application request"
            );
            Status::invalid_argument(format!("Invalid request data: {}", e))
        })?;

        // Call service layer with permission validation
        match self.state.application_service.update_application_with_permission(
            &grpc_request.id,
            update_request,
            &grpc_request.requesting_user_id,
            &grpc_request.requesting_user_role,
        ).await {
            Ok(Some(application_response)) => {
                info!(
                    request_id = %request_id,
                    application_id = %grpc_request.id,
                    "Successfully updated application via gRPC with permission validation"
                );
                
                let grpc_response = application_response_to_grpc(&application_response);
                Ok(Response::new(grpc_response))
            }
            Ok(None) => {
                warn!(
                    request_id = %request_id,
                    application_id = %grpc_request.id,
                    "Application not found for update via gRPC"
                );
                Err(Status::not_found(format!("Application with ID {} not found", grpc_request.id)))
            }
            Err(e) => {
                error!(
                    request_id = %request_id,
                    application_id = %grpc_request.id,
                    error = %e,
                    "Failed to update application via gRPC"
                );
                Err(map_app_error_to_status(e))
            }
        }
    }

    /// Change application status (partial update)
    /// 
    /// Updates only the status and feedback of an application. This is commonly
    /// used for workflow state transitions (draft -> pending -> approved/rejected).
    /// Now includes permission validation to prevent security vulnerabilities.
    async fn change_application_status(
        &self,
        request: Request<proto::ChangeStatusRequest>,
    ) -> Result<Response<proto::StatusChangeResponse>, Status> {
        let (user_agent, request_id) = extract_request_metadata(&request);
        let grpc_request = request.into_inner();
        
        info!(
            request_id = %request_id,
            user_agent = %user_agent,
            application_id = %grpc_request.id,
            status = ?grpc_request.status,
            requesting_user_id = %grpc_request.requesting_user_id,
            requesting_user_role = %grpc_request.requesting_user_role,
            "Received gRPC change application status request"
        );

        // Validate required permission fields
        if grpc_request.requesting_user_id.is_empty() || grpc_request.requesting_user_role.is_empty() {
            return Err(Status::invalid_argument("requesting_user_id and requesting_user_role are required"));
        }

        // Convert status
        let status = grpc_to_application_status(grpc_request.status)?;

        // Create internal request DTO
        let status_request = ChangeStatusRequest {
            status: status.clone(),
            feedback: grpc_request.feedback,
        };

        // Validate the request using existing validation rules
        status_request.validate().map_err(|e| {
            warn!(
                request_id = %request_id,
                error = %e,
                "Validation failed for change status request"
            );
            Status::invalid_argument(format!("Invalid request data: {}", e))
        })?;

        // Call service layer with permission validation
        match self.state.application_service.update_application_status_with_permission(
            &grpc_request.id,
            status_request,
            &grpc_request.requesting_user_id,
            &grpc_request.requesting_user_role,
        ).await {
            Ok(Some(application_response)) => {
                info!(
                    request_id = %request_id,
                    application_id = %grpc_request.id,
                    new_status = ?application_response.status,
                    "Successfully changed application status via gRPC with permission validation"
                );
                
                let response = proto::StatusChangeResponse {
                    status: application_status_to_grpc(&application_response.status) as i32,
                    status_display: application_response.status_display,
                    feedback: application_response.feedback,
                };
                
                Ok(Response::new(response))
            }
            Ok(None) => {
                warn!(
                    request_id = %request_id,
                    application_id = %grpc_request.id,
                    "Application not found for status change via gRPC"
                );
                Err(Status::not_found(format!("Application with ID {} not found", grpc_request.id)))
            }
            Err(e) => {
                error!(
                    request_id = %request_id,
                    application_id = %grpc_request.id,
                    error = %e,
                    "Failed to change application status via gRPC"
                );
                Err(map_app_error_to_status(e))
            }
        }
    }

    /// Delete an application
    /// 
    /// Permanently removes an application and all its associated data from the system.
    /// This operation is irreversible.
    /// Now includes permission validation to prevent security vulnerabilities.
    async fn delete_application(
        &self,
        request: Request<proto::DeleteApplicationRequest>,
    ) -> Result<Response<proto::DeleteApplicationResponse>, Status> {
        let (user_agent, request_id) = extract_request_metadata(&request);
        let grpc_request = request.into_inner();
        
        info!(
            request_id = %request_id,
            user_agent = %user_agent,
            application_id = %grpc_request.id,
            requesting_user_id = %grpc_request.requesting_user_id,
            requesting_user_role = %grpc_request.requesting_user_role,
            "Received gRPC delete application request"
        );

        // Validate required permission fields
        if grpc_request.requesting_user_id.is_empty() || grpc_request.requesting_user_role.is_empty() {
            return Err(Status::invalid_argument("requesting_user_id and requesting_user_role are required"));
        }

        // Call service layer with permission validation
        match self.state.application_service.delete_application_with_permission(
            &grpc_request.id,
            &grpc_request.requesting_user_id,
            &grpc_request.requesting_user_role,
        ).await {
            Ok(true) => {
                info!(
                    request_id = %request_id,
                    application_id = %grpc_request.id,
                    "Successfully deleted application via gRPC with permission validation"
                );
                
                let response = proto::DeleteApplicationResponse {
                    success: true,
                    message: format!("Application {} successfully deleted", grpc_request.id),
                };
                
                Ok(Response::new(response))
            }
            Ok(false) => {
                warn!(
                    request_id = %request_id,
                    application_id = %grpc_request.id,
                    "Application not found for deletion via gRPC"
                );
                Err(Status::not_found(format!("Application with ID {} not found", grpc_request.id)))
            }
            Err(e) => {
                error!(
                    request_id = %request_id,
                    application_id = %grpc_request.id,
                    error = %e,
                    "Failed to delete application via gRPC"
                );
                Err(map_app_error_to_status(e))
            }
        }
    }

    /// List applications with filtering and pagination
    /// 
    /// This method is implemented in the query_grpc_handlers module to maintain
    /// separation of concerns. It provides the same filtering and pagination
    /// capabilities as the REST API.
    async fn list_applications(
        &self,
        request: Request<proto::ListApplicationsRequest>,
    ) -> Result<Response<proto::PaginatedApplicationResponse>, Status> {
        // Delegate to query handlers for consistency
        let query_service = super::query_grpc_handlers::QueryGrpcService::new(self.state.clone());
        query_service.list_applications_impl(request).await
    }
}
