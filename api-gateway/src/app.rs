use axum::Router;
use sqlx::postgres::PgPoolOptions;
use std::sync::Arc;
use crate::service::auth_service::AuthService;
use crate::routes::auth_routes;

pub async fn create_app(database_url: &str, jwt_secret: &str, access_token_expiry_minutes: i64, refresh_token_expiry_minutes: i64) -> Router {
    let pool = Arc::new(
        PgPoolOptions::new()
            .max_connections(5)
            .connect(database_url)
            .await
            .expect("Failed to connect to DB"),
    );
    let auth_service = Arc::new(AuthService::new(pool, jwt_secret.to_owned(), access_token_expiry_minutes, refresh_token_expiry_minutes));
    Router::new()
        .nest("/auth", auth_routes::auth_routes(auth_service))
} 