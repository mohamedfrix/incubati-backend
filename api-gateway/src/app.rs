use axum::Router;
use sqlx::postgres::PgPoolOptions;
use std::sync::Arc;
use crate::service::auth_service::AuthService;
use crate::routes::{auth_routes, application_routes};
use crate::config::Config;
use api_gateway::grpc_client::GrpcApplicationClient;
use axum::Extension;

pub async fn create_app(config: &Config) -> Router {
    let pool = Arc::new(
        PgPoolOptions::new()
            .max_connections(5)
            .connect(&config.database_url)
            .await
            .expect("Failed to connect to DB"),
    );
    
    let auth_service = Arc::new(AuthService::new(
        pool, 
        config.jwt_secret.clone(), 
        config.access_token_expiry_minutes, 
        config.refresh_token_expiry_minutes
    ));
    
    // Connect to gRPC service with retry logic
    let grpc_client = Arc::new(
        connect_to_grpc_with_retry(&config.grpc_application_service_url).await
            .expect("Failed to connect to gRPC ApplicationService after retries"),
    );
    
    Router::new()
        .nest("/auth", auth_routes::auth_routes(auth_service.clone()))
        .nest("/applications", 
            application_routes::application_routes()
                .layer(Extension(grpc_client.clone()))
                .layer(Extension(auth_service.clone()))
        )
        .nest("/documents", 
            application_routes::documents_router()
                .layer(Extension(grpc_client))
                .layer(Extension(auth_service))
        )
}

/// Connect to gRPC service with retry logic
async fn connect_to_grpc_with_retry(grpc_url: &str) -> Result<GrpcApplicationClient, Box<dyn std::error::Error + Send + Sync>> {
    use tokio::time::{sleep, Duration};
    use tracing::{info, warn};
    
    let max_retries = 10;
    let mut retry_count = 0;
    
    info!("Attempting to connect to gRPC service at: {}", grpc_url);
    
    loop {
        match GrpcApplicationClient::connect(grpc_url).await {
            Ok(client) => {
                info!("Successfully connected to gRPC ApplicationService at {}", grpc_url);
                return Ok(client);
            }
            Err(e) => {
                retry_count += 1;
                if retry_count >= max_retries {
                    return Err(format!("Failed to connect to gRPC service at {} after {} retries: {}", grpc_url, max_retries, e).into());
                }
                
                warn!("Failed to connect to gRPC service at {} (attempt {}/{}): {}. Retrying in 2 seconds...", grpc_url, retry_count, max_retries, e);
                sleep(Duration::from_secs(2)).await;
            }
        }
    }
} 