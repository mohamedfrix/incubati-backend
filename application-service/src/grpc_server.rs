/// gRPC server implementation
/// 
/// This module sets up and configures the gRPC server with all service implementations.
/// It handles:
/// - Server configuration and middleware
/// - Service registration
/// - Health checks and reflection (optional)
/// - Graceful shutdown
/// 
/// The server can run alongside the HTTP/REST server or independently.

use std::sync::Arc;
use tonic::{transport::Server, Request, Status};
use tracing::{info, error};

use crate::service::{ApplicationService, QueryService, DocumentService};
use crate::grpc_handlers::{
    GrpcServiceState,
    application_grpc_handlers::ApplicationGrpcService,
    document_grpc_handlers::DocumentGrpcService,
    conversions::proto::{
        application_service_server::ApplicationServiceServer,
        document_service_server::DocumentServiceServer,
    },
};

/// gRPC server configuration
#[derive(Debug, Clone)]
pub struct GrpcServerConfig {
    pub host: String,
    pub port: u16,
    pub enable_reflection: bool,
    pub enable_health_check: bool,
    pub max_message_size: Option<usize>,
    pub timeout_seconds: Option<u64>,
}

impl Default for GrpcServerConfig {
    fn default() -> Self {
        Self {
            host: "0.0.0.0".to_string(),
            port: 50051,
            enable_reflection: true,  // Useful for development with tools like grpcurl
            enable_health_check: true,
            max_message_size: Some(4 * 1024 * 1024), // 4MB default
            timeout_seconds: Some(30),
        }
    }
}

/// gRPC server implementation
pub struct GrpcServer {
    config: GrpcServerConfig,
    application_service: Arc<ApplicationService>,
    query_service: Arc<QueryService>,
    document_service: Arc<DocumentService>,
}

impl GrpcServer {
    /// Create a new gRPC server
    pub fn new(
        config: GrpcServerConfig,
        application_service: Arc<ApplicationService>,
        query_service: Arc<QueryService>,
        document_service: Arc<DocumentService>,
    ) -> Self {
        Self {
            config,
            application_service,
            query_service,
            document_service,
        }
    }

    /// Start the gRPC server
    /// 
    /// This method configures and starts the gRPC server with all services.
    /// It blocks until the server is shut down.
    pub async fn serve(self) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        let addr = format!("{}:{}", self.config.host, self.config.port).parse()?;
        
        info!("Starting gRPC server on {}", addr);

        // Create shared state for gRPC services
        let grpc_state = GrpcServiceState::new(
            self.application_service.clone(),
            self.query_service.clone(),
            self.document_service.clone(),
        );

        // Create service implementations
        let application_service = ApplicationGrpcService::new(grpc_state.clone());
        let document_service = DocumentGrpcService::new(grpc_state);

        // Create server builder
        let server_builder = Server::builder()
            .timeout(std::time::Duration::from_secs(
                self.config.timeout_seconds.unwrap_or(30)
            ));

        // Add interceptor for logging and error handling
        let mut server_builder = server_builder
            .layer(tonic::service::interceptor(request_interceptor));

        // Build router with services
        let mut router = server_builder
            .add_service(ApplicationServiceServer::new(application_service))
            .add_service(DocumentServiceServer::new(document_service));

        // Add reflection service if enabled (useful for development)
        if self.config.enable_reflection {
            // For now, we'll skip reflection until we properly generate the descriptor set
            // You can add it back once you set up proper descriptor generation in build.rs
            info!("gRPC reflection skipped (descriptor set not available)");
        }

        // Add health check service if enabled
        if self.config.enable_health_check {
            let (mut health_reporter, health_service) = tonic_health::server::health_reporter();
            health_reporter
                .set_serving::<ApplicationServiceServer<ApplicationGrpcService>>()
                .await;
            health_reporter
                .set_serving::<DocumentServiceServer<DocumentGrpcService>>()
                .await;
            
            router = router.add_service(health_service);
            info!("gRPC health check enabled");
        }

        info!("gRPC server configured and starting...");

        // Start serving
        router
            .serve_with_shutdown(addr, self.shutdown_signal())
            .await?;

        info!("gRPC server stopped");
        Ok(())
    }

    /// Handle graceful shutdown
    async fn shutdown_signal(&self) {
        // Wait for the CTRL+C signal
        tokio::signal::ctrl_c()
            .await
            .expect("Failed to install CTRL+C signal handler");
        
        info!("Received shutdown signal, stopping gRPC server...");
    }
}

/// Request interceptor for logging and error handling
/// 
/// This interceptor logs all incoming requests and adds common error handling.
/// It can be extended to add authentication, rate limiting, etc.
fn request_interceptor(req: Request<()>) -> Result<Request<()>, Status> {
    let metadata = req.metadata();
    
    // Extract useful information for logging
    let user_agent = metadata
        .get("user-agent")
        .and_then(|v| v.to_str().ok())
        .unwrap_or("unknown");
    
    let content_type = metadata
        .get("content-type")
        .and_then(|v| v.to_str().ok())
        .unwrap_or("unknown");

    // Log the request (you might want to be more selective about what you log)
    info!(
        user_agent = %user_agent,
        content_type = %content_type,
        "Incoming gRPC request"
    );

    // You can add authentication, authorization, or other middleware logic here
    // For example:
    // if !authenticate_request(&req) {
    //     return Err(Status::unauthenticated("Invalid credentials"));
    // }

    Ok(req)
}

/// Utility function to create a gRPC server with default configuration
pub async fn create_grpc_server(
    application_service: Arc<ApplicationService>,
    query_service: Arc<QueryService>,
    document_service: Arc<DocumentService>,
) -> GrpcServer {
    let config = GrpcServerConfig::default();
    GrpcServer::new(config, application_service, query_service, document_service)
}

/// Utility function to create a gRPC server with custom configuration
pub async fn create_grpc_server_with_config(
    config: GrpcServerConfig,
    application_service: Arc<ApplicationService>,
    query_service: Arc<QueryService>,
    document_service: Arc<DocumentService>,
) -> GrpcServer {
    GrpcServer::new(config, application_service, query_service, document_service)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_config() {
        let config = GrpcServerConfig::default();
        assert_eq!(config.host, "0.0.0.0");
        assert_eq!(config.port, 50051);
        assert!(config.enable_reflection);
        assert!(config.enable_health_check);
        assert_eq!(config.max_message_size, Some(4 * 1024 * 1024));
        assert_eq!(config.timeout_seconds, Some(30));
    }

    #[test] 
    fn test_server_address_parsing() {
        let config = GrpcServerConfig::default();
        let addr = format!("{}:{}", config.host, config.port);
        let parsed = addr.parse::<std::net::SocketAddr>();
        assert!(parsed.is_ok());
    }
}
