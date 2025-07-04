/// gRPC handlers for document management operations
/// 
/// This module implements document-related gRPC operations such as:
/// - Uploading documents to applications
/// - Retrieving document metadata
/// - Downloading document content
/// - Deleting documents
/// 
/// Note: Document upload/download via gRPC requires special handling for binary data.
/// For large files, it's recommended to use streaming RPCs or provide HTTP endpoints
/// for actual file transfer while using gRPC for metadata operations.

use tonic::{Request, Response, Status};
use tracing::{info, warn, error, debug};
use validator::Validate;
use tokio_stream::Stream;
use bytes;

use super::{GrpcServiceState, map_app_error_to_status, extract_request_metadata};
use super::conversions::{proto, *, document_download_response_to_grpc};
use crate::dto::*;

/// gRPC service implementation for document operations
pub struct DocumentGrpcService {
    state: GrpcServiceState,
}

impl DocumentGrpcService {
    /// Create a new DocumentGrpcService
    pub fn new(state: GrpcServiceState) -> Self {
        Self { state }
    }
}

#[tonic::async_trait]
impl super::conversions::proto::document_service_server::DocumentService for DocumentGrpcService {

    /// Get documents for a specific application
    /// 
    /// Retrieves metadata for all documents associated with an application.
    /// This is useful for listing available documents before downloading.
    async fn get_application_documents(
        &self,
        request: Request<proto::GetApplicationRequest>,
    ) -> Result<Response<proto::ApplicationDocumentsResponse>, Status> {
        let (user_agent, request_id) = extract_request_metadata(&request);
        let grpc_request = request.into_inner();
        
        info!(
            request_id = %request_id,
            user_agent = %user_agent,
            application_id = %grpc_request.id,
            "Received gRPC get application documents request"
        );

        // Parse and validate UUID
        let application_id = parse_uuid(&grpc_request.id)?;

        // Call service layer to get documents
        match self.state.document_service.get_application_documents(application_id).await {
            Ok(documents) => {
                info!(
                    request_id = %request_id,
                    application_id = %grpc_request.id,
                    document_count = documents.len(),
                    "Successfully retrieved application documents via gRPC"
                );
                
                let grpc_documents: Vec<proto::DocumentResponse> = documents
                    .iter()
                    .map(document_response_to_grpc)
                    .collect();

                let response = proto::ApplicationDocumentsResponse {
                    application_id: grpc_request.id.clone(),
                    documents: grpc_documents,
                };
                
                Ok(Response::new(response))
            }
            Err(e) => {
                error!(
                    request_id = %request_id,
                    application_id = %grpc_request.id,
                    error = %e,
                    "Failed to retrieve application documents via gRPC"
                );
                Err(map_app_error_to_status(e))
            }
        }
    }

    /// Get document metadata by ID
    /// 
    /// Retrieves metadata for a specific document without downloading the content.
    /// Useful for validation and display purposes.
    /// Now includes permission validation to prevent security vulnerabilities.
    async fn get_document(
        &self,
        request: Request<proto::GetDocumentRequest>,
    ) -> Result<Response<proto::DocumentResponse>, Status> {
        let (user_agent, request_id) = extract_request_metadata(&request);
        let grpc_request = request.into_inner();
        
        info!(
            request_id = %request_id,
            user_agent = %user_agent,
            document_id = %grpc_request.id,
            requesting_user_id = %grpc_request.requesting_user_id,
            requesting_user_role = %grpc_request.requesting_user_role,
            "Received gRPC get document request"
        );

        // Validate required permission fields
        if grpc_request.requesting_user_id.is_empty() || grpc_request.requesting_user_role.is_empty() {
            return Err(Status::invalid_argument("requesting_user_id and requesting_user_role are required"));
        }

        // Call service layer with permission validation
        match self.state.document_service.get_document_with_permission(
            &grpc_request.id,
            &grpc_request.requesting_user_id,
            &grpc_request.requesting_user_role,
        ).await {
            Ok(document) => {
                info!(
                    request_id = %request_id,
                    document_id = %grpc_request.id,
                    "Successfully retrieved document via gRPC with permission validation"
                );
                
                let grpc_response = document_response_to_grpc(&document);
                Ok(Response::new(grpc_response))
            }
            Err(e) => {
                error!(
                    request_id = %request_id,
                    document_id = %grpc_request.id,
                    error = %e,
                    "Failed to retrieve document via gRPC"
                );
                Err(map_app_error_to_status(e))
            }
        }
    }

    /// Create document metadata (for upload preparation)
    /// 
    /// Creates a document record with metadata. The actual file content
    /// should be uploaded separately through HTTP endpoints for better
    /// performance with large files.
    /// Now includes permission validation to prevent security vulnerabilities.
    async fn create_document(
        &self,
        request: Request<proto::CreateDocumentRequest>,
    ) -> Result<Response<proto::DocumentResponse>, Status> {
        let (user_agent, request_id) = extract_request_metadata(&request);
        let grpc_request = request.into_inner();
        
        info!(
            request_id = %request_id,
            user_agent = %user_agent,
            application_id = %grpc_request.application_id,
            original_name = %grpc_request.original_name,
            requesting_user_id = %grpc_request.requesting_user_id,
            requesting_user_role = %grpc_request.requesting_user_role,
            "Received gRPC create document request"
        );

        // Validate required permission fields
        if grpc_request.requesting_user_id.is_empty() || grpc_request.requesting_user_role.is_empty() {
            return Err(Status::invalid_argument("requesting_user_id and requesting_user_role are required"));
        }

        // Debug logging for detailed analysis
        debug!(
            request_id = %request_id,
            application_id_raw = %grpc_request.application_id,
            application_id_len = grpc_request.application_id.len(),
            application_id_bytes = ?grpc_request.application_id.as_bytes(),
            original_name_raw = %grpc_request.original_name,
            original_name_len = grpc_request.original_name.len(),
            file_content_len = grpc_request.file_content.len(),
            file_content_first_10_bytes = ?&grpc_request.file_content[..std::cmp::min(10, grpc_request.file_content.len())],
            content_type_raw = %grpc_request.content_type,
            content_type_len = grpc_request.content_type.len(),
            document_type_raw = %grpc_request.document_type,
            document_type_len = grpc_request.document_type.len(),
            "Debug: Raw field values received"
        );

        // Validate required fields
        debug!(
            request_id = %request_id,
            original_name_trimmed = %grpc_request.original_name.trim(),
            original_name_is_empty = grpc_request.original_name.trim().is_empty(),
            "Debug: Validating original name"
        );
        if grpc_request.original_name.trim().is_empty() {
            error!(
                request_id = %request_id,
                "Debug: Original name validation failed - field is empty after trim"
            );
            return Err(Status::invalid_argument("Original name is required"));
        }

        debug!(
            request_id = %request_id,
            file_content_length = grpc_request.file_content.len(),
            file_content_is_empty = grpc_request.file_content.is_empty(),
            "Debug: Validating file content"
        );
        if grpc_request.file_content.is_empty() {
            error!(
                request_id = %request_id,
                "Debug: File content validation failed - Vec<u8> is empty"
            );
            return Err(Status::invalid_argument("File content is required"));
        }

        // Prepare file data
        let file_content = bytes::Bytes::from(grpc_request.file_content);
        debug!(
            request_id = %request_id,
            bytes_length = file_content.len(),
            "Debug: Converted Vec<u8> to Bytes"
        );
        
        let content_type = if grpc_request.content_type.trim().is_empty() {
            debug!(
                request_id = %request_id,
                "Debug: Content type is empty, using default: application/octet-stream"
            );
            "application/octet-stream".to_string()
        } else {
            debug!(
                request_id = %request_id,
                content_type = %grpc_request.content_type,
                "Debug: Using provided content type"
            );
            grpc_request.content_type
        };
        let document_type = if grpc_request.document_type.trim().is_empty() {
            debug!(
                request_id = %request_id,
                "Debug: Document type is empty, using default: general"
            );
            "general".to_string()
        } else {
            debug!(
                request_id = %request_id,
                document_type = %grpc_request.document_type,
                "Debug: Using provided document type"
            );
            grpc_request.document_type
        };

        debug!(
            request_id = %request_id,
            application_id = %grpc_request.application_id,
            original_name = %grpc_request.original_name,
            content_type = %content_type,
            document_type = %document_type,
            file_size = file_content.len(),
            "Debug: About to call document service upload_document"
        );

        // Call service layer to upload document with permission validation
        match self.state.document_service
            .upload_document_with_permission(
                &grpc_request.application_id,
                grpc_request.original_name.clone(),
                file_content,
                content_type,
                document_type,
                &grpc_request.requesting_user_id,
                &grpc_request.requesting_user_role,
            )
            .await
        {
            Ok(document) => {
                info!(
                    request_id = %request_id,
                    application_id = %grpc_request.application_id,
                    document_id = %document.id,
                    original_name = %grpc_request.original_name,
                    "Successfully created document via gRPC"
                );
                
                let grpc_response = document_response_to_grpc(&document);
                Ok(Response::new(grpc_response))
            }
            Err(e) => {
                error!(
                    request_id = %request_id,
                    application_id = %grpc_request.application_id,
                    original_name = %grpc_request.original_name,
                    error = %e,
                    "Failed to create document via gRPC"
                );
                Err(map_app_error_to_status(e))
            }
        }
    }

    /// Delete a document
    /// 
    /// Removes both the document metadata and the associated file from storage.
    /// This operation is irreversible.
    /// Now includes permission validation to prevent security vulnerabilities.
    async fn delete_document(
        &self,
        request: Request<proto::DeleteDocumentRequest>,
    ) -> Result<Response<proto::DeleteDocumentResponse>, Status> {
        let (user_agent, request_id) = extract_request_metadata(&request);
        let grpc_request = request.into_inner();
        
        info!(
            request_id = %request_id,
            user_agent = %user_agent,
            document_id = %grpc_request.id,
            requesting_user_id = %grpc_request.requesting_user_id,
            requesting_user_role = %grpc_request.requesting_user_role,
            "Received gRPC delete document request"
        );

        // Validate required permission fields
        if grpc_request.requesting_user_id.is_empty() || grpc_request.requesting_user_role.is_empty() {
            return Err(Status::invalid_argument("requesting_user_id and requesting_user_role are required"));
        }

        // Call service layer with permission validation
        match self.state.document_service.delete_document_with_permission(
            &grpc_request.id,
            &grpc_request.requesting_user_id,
            &grpc_request.requesting_user_role,
        ).await {
            Ok(()) => {
                info!(
                    request_id = %request_id,
                    document_id = %grpc_request.id,
                    "Successfully deleted document via gRPC with permission validation"
                );
                
                let response = proto::DeleteDocumentResponse {
                    success: true,
                    message: format!("Document {} successfully deleted", grpc_request.id),
                };
                
                Ok(Response::new(response))
            }
            Err(e) => {
                error!(
                    request_id = %request_id,
                    document_id = %grpc_request.id,
                    error = %e,
                    "Failed to delete document via gRPC"
                );
                Err(map_app_error_to_status(e))
            }
        }
    }

    /// Get document download URL (for future implementation)
    /// 
    /// Instead of streaming large files through gRPC, this method could
    /// return a signed URL for direct HTTP download. This is more efficient
    /// for large files and reduces gRPC server load.
    #[allow(dead_code)]
    async fn get_document_download_url(
        &self,
        request: Request<proto::GetDocumentRequest>,
    ) -> Result<Response<proto::DocumentDownloadUrlResponse>, Status> {
        let (user_agent, request_id) = extract_request_metadata(&request);
        let grpc_request = request.into_inner();
        
        info!(
            request_id = %request_id,
            user_agent = %user_agent,
            document_id = %grpc_request.id,
            requesting_user_id = %grpc_request.requesting_user_id,
            requesting_user_role = %grpc_request.requesting_user_role,
            "Received gRPC get document download URL request"
        );

        // Validate required permission fields
        if grpc_request.requesting_user_id.is_empty() || grpc_request.requesting_user_role.is_empty() {
            return Err(Status::invalid_argument("requesting_user_id and requesting_user_role are required"));
        }

        // Call service layer with permission validation
        // Default expiry of 1 hour (3600 seconds) for download URLs
        match self.state.document_service.generate_download_url_with_permission(
            &grpc_request.id,
            &grpc_request.requesting_user_id,
            &grpc_request.requesting_user_role,
            Some(3600), // 1 hour expiry
        ).await {
            Ok(download_response) => {
                info!(
                    request_id = %request_id,
                    document_id = %grpc_request.id,
                    download_url = %download_response.download_url,
                    "Successfully generated document download URL via gRPC"
                );
                
                let grpc_response = document_download_response_to_grpc(&download_response);
                Ok(Response::new(grpc_response))
            }
            Err(e) => {
                error!(
                    request_id = %request_id,
                    document_id = %grpc_request.id,
                    error = %e,
                    "Failed to generate document download URL via gRPC"
                );
                Err(map_app_error_to_status(e))
            }
        }
    }

    /// Upload document content via streaming (for future implementation)
    /// 
    /// This would allow uploading document content directly through gRPC
    /// using streaming. However, for large files, HTTP uploads are typically
    /// more efficient.
    #[allow(dead_code)]
    async fn upload_document_stream(
        &self,
        _request: Request<tonic::Streaming<proto::DocumentChunk>>,
    ) -> Result<Response<proto::DocumentResponse>, Status> {
        // Implementation placeholder for future enhancement
        Err(Status::unimplemented("Streaming upload not yet implemented"))
    }

    /// Download document content via streaming (for future implementation)
    /// 
    /// This would allow downloading document content directly through gRPC
    /// using streaming. However, for large files, HTTP downloads are typically
    /// more efficient.
    type DownloadDocumentStreamStream = std::pin::Pin<Box<dyn tokio_stream::Stream<Item = Result<proto::DocumentChunk, tonic::Status>> + Send + 'static>>;
    
    #[allow(dead_code)]
    async fn download_document_stream(
        &self,
        _request: Request<proto::GetDocumentRequest>,
    ) -> Result<Response<Self::DownloadDocumentStreamStream>, Status> {
        // Implementation placeholder for future enhancement
        Err(Status::unimplemented("Streaming download not yet implemented"))
    }
}

// Additional proto message definitions would be needed for the future implementations:
// 
// message DocumentDownloadUrlResponse {
//     string download_url = 1;
//     google.protobuf.Timestamp expires_at = 2;
// }
// 
// message DocumentChunk {
//     bytes data = 1;
//     int32 chunk_number = 2;
//     bool is_last_chunk = 3;
// }
// 
// message ApplicationDocumentsResponse {
//     string application_id = 1;
//     repeated DocumentResponse documents = 2;
// }
// 
// message CreateDocumentRequest {
//     string application_id = 1;
//     string title = 2;
// }
// 
// message GetDocumentRequest {
//     string id = 1;
// }
// 
// message DeleteDocumentRequest {
//     string id = 1;
// }
// 
// message DeleteDocumentResponse {
//     bool success = 1;
//     string message = 2;
// }
