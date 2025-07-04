/// gRPC handlers module
/// 
/// This module contains all gRPC service implementations following the same
/// architectural pattern as the REST handlers. Each handler:
/// 1. Validates the incoming request data
/// 2. Converts gRPC types to internal DTOs
/// 3. Calls the appropriate service layer method
/// 4. Converts the response back to gRPC types
/// 5. Handles errors appropriately

pub mod application_grpc_handlers;
pub mod query_grpc_handlers;
pub mod document_grpc_handlers;
pub mod conversions;

use std::sync::Arc;
use tonic::{Request, Response, Status};

use crate::service::{ApplicationService, QueryService, DocumentService};

/// gRPC service state containing shared services
/// 
/// This mirrors the AppState used by REST handlers but adapted for gRPC
#[derive(Clone)]
pub struct GrpcServiceState {
    pub application_service: Arc<ApplicationService>,
    pub query_service: Arc<QueryService>,
    pub document_service: Arc<DocumentService>,
}

impl GrpcServiceState {
    /// Create a new gRPC service state
    pub fn new(
        application_service: Arc<ApplicationService>,
        query_service: Arc<QueryService>,
        document_service: Arc<DocumentService>,
    ) -> Self {
        Self {
            application_service,
            query_service,
            document_service,
        }
    }
}

/// Common error handling for gRPC services
/// 
/// Converts internal application errors to appropriate gRPC status codes
pub fn map_app_error_to_status(error: crate::errors::AppError) -> Status {
    use crate::errors::AppError;
    
    match error {
        AppError::NotFound(msg) => Status::not_found(msg),
        AppError::Validation(msg) => Status::invalid_argument(msg),
        AppError::BadRequest(msg) => Status::invalid_argument(msg),
        AppError::Internal(err) => Status::internal(format!("Internal error: {}", err)),
        AppError::Database(err) => Status::internal(format!("Database error: {}", err)),
        AppError::Unauthorized(msg) => Status::unauthenticated(msg),
        AppError::Conflict(msg) => Status::already_exists(msg),
        AppError::FileError(msg) => Status::internal(format!("File error: {}", msg)),
        AppError::PermissionDenied(msg) => Status::permission_denied(msg),
        AppError::InvalidRole { role } => Status::invalid_argument(format!("Invalid user role: {}", role)),
    }
}

/// Extract metadata from gRPC request for logging and tracing
pub fn extract_request_metadata<T>(request: &Request<T>) -> (String, String) {
    let metadata = request.metadata();
    
    let user_agent = metadata
        .get("user-agent")
        .and_then(|v| v.to_str().ok())
        .unwrap_or("unknown")
        .to_string();
        
    let request_id = metadata
        .get("x-request-id")
        .and_then(|v| v.to_str().ok())
        .unwrap_or("unknown")
        .to_string();
        
    (user_agent, request_id)
}
