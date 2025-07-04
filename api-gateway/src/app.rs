use axum::Router;
use sqlx::postgres::PgPoolOptions;
use std::sync::Arc;
use crate::service::auth_service::AuthService;
use crate::routes::{auth_routes, application_routes};
use api_gateway::grpc_client::GrpcApplicationClient;
use axum::Extension;

pub async fn create_app(database_url: &str, jwt_secret: &str, access_token_expiry_minutes: i64, refresh_token_expiry_minutes: i64) -> Router {
    let pool = Arc::new(
        PgPoolOptions::new()
            .max_connections(5)
            .connect(database_url)
            .await
            .expect("Failed to connect to DB"),
    );
    let auth_service = Arc::new(AuthService::new(pool, jwt_secret.to_owned(), access_token_expiry_minutes, refresh_token_expiry_minutes));
    let grpc_client = Arc::new(
        GrpcApplicationClient::connect("http://127.0.0.1:50051")
            .await
            .expect("Failed to connect to gRPC ApplicationService"),
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