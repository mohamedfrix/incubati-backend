pub mod config;
pub mod dto;
pub mod errors;
pub mod handlers;
pub mod grpc_handlers;
pub mod grpc_server;
pub mod models;
pub mod repository;
pub mod service;
pub mod utils;

use std::sync::Arc;
use axum::Router;
use sqlx::PgPool;
use tower::ServiceBuilder;
use tower_http::{cors::CorsLayer, trace::TraceLayer};
use tracing::{info, error};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use config::Config;
use handlers::{AppState, create_router, create_utility_router};
use repository::{PgApplicationRepository, PgDocumentRepository};
use service::{ApplicationService, QueryService, DocumentService};
use utils::MinioClient;
use grpc_server::{GrpcServer, GrpcServerConfig};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    // Initialize tracing
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "application_service=debug,tower_http=debug".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    info!("Starting Application Service...");

    // Load configuration
    let config = Config::from_env()?;
    info!("Configuration loaded successfully");

    // Connect to database
    info!("Connecting to database: {}", config.database_url);
    let pool = PgPool::connect(&config.database_url).await?;
    info!("Database connection established");

    // Run migrations
    info!("Running database migrations...");
    sqlx::migrate!("./migrations").run(&pool).await?;
    info!("Database migrations completed");

    // Initialize Minio client
    info!("Initializing Minio client...");
    let minio_client = Arc::new(MinioClient::new(config.minio.clone())?);
    
    // Initialize Minio buckets
    minio_client.initialize_buckets().await?;
    info!("Minio client initialized successfully");

    // Setup dependencies
    let application_repository = Arc::new(PgApplicationRepository::new(pool.clone()));
    let document_repository = Arc::new(PgDocumentRepository::new(pool));
    let application_service = Arc::new(ApplicationService::new(application_repository.clone(), document_repository.clone()));
    let query_service = Arc::new(QueryService::new(application_repository.clone()));
    let document_service = Arc::new(DocumentService::new(application_repository, document_repository, minio_client));

    // Create application state for REST API
    let app_state = AppState {
        application_service: application_service.clone(),
        query_service: query_service.clone(),
        document_service: document_service.clone(),
    };

    // Setup CORS
    let cors = CorsLayer::new()
        .allow_origin(tower_http::cors::Any)
        .allow_methods(tower_http::cors::Any)
        .allow_headers(tower_http::cors::Any);

    // Create the main application router for REST API
    let app = Router::new()
        .nest("/api", create_router(app_state))
        .merge(create_utility_router())
        .layer(
            ServiceBuilder::new()
                .layer(TraceLayer::new_for_http())
                .layer(cors)
        );

    // Create gRPC server configuration
    let grpc_config = GrpcServerConfig {
        host: "0.0.0.0".to_string(),
        port: config.grpc_port,
        enable_reflection: true,
        enable_health_check: true,
        max_message_size: Some(4 * 1024 * 1024),
        timeout_seconds: Some(30),
    };

    // Create gRPC server
    let grpc_server = GrpcServer::new(
        grpc_config,
        application_service,
        query_service,
        document_service,
    );

    // Start both servers concurrently
    let http_addr = format!("{}:{}", config.server_host, config.server_port);
    let grpc_addr = format!("0.0.0.0:{}", config.grpc_port);
    
    info!("Starting HTTP server on {}", http_addr);
    info!("Starting gRPC server on {}", grpc_addr);

    // Run both servers concurrently
    let http_server = async {
        let listener = tokio::net::TcpListener::bind(&http_addr).await?;
        info!("HTTP server listening on {}", http_addr);
        axum::serve(listener, app).await
    };

    let grpc_server_task = async {
        grpc_server.serve().await
    };

    // Use tokio::select! to run both servers concurrently
    tokio::select! {
        result = http_server => {
            if let Err(e) = result {
                error!("HTTP server error: {}", e);
                return Err(e.into());
            }
        }
        result = grpc_server_task => {
            if let Err(e) = result {
                error!("gRPC server error: {}", e);
                return Err(e);
            }
        }
    }

    info!("Application Service stopped");
    Ok(())
}
