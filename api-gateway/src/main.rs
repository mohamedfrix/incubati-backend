mod config;
mod app;
mod models;
mod repository;
mod service;
mod dto;
mod routes;
mod utils;
mod middleware;

use crate::config::Config;
use axum::serve;
use tokio::net::TcpListener;
use tracing_subscriber::fmt;
use tracing_subscriber::EnvFilter;
use tracing::{info, debug, warn, error};
use std::net::SocketAddr;

#[tokio::main]
async fn main() {
    // Initialize logging with debug as default
    fmt()
        .with_env_filter(EnvFilter::try_new("debug").unwrap_or_else(|_| EnvFilter::new("debug")))
        .init();

    debug!("Logger initialized at debug level");

    // Load configuration from .env
    let config = Config::from_env();
    debug!(?config, "Loaded configuration");

    // Build the Axum app with all routes
    debug!("Connecting to database...");
    let app = app::create_app(
        &config.database_url,
        &config.jwt_secret,
        config.access_token_expiry_minutes,
        config.refresh_token_expiry_minutes,
    ).await;
    info!("App and routes initialized");

    // Start the server
    let addr = SocketAddr::from(([0, 0, 0, 0], 4000));
    debug!(?addr, "Binding TCP listener");
    let listener = TcpListener::bind(addr).await.unwrap();
    info!("Auth Service running on {}", addr);
    serve(listener, app).await.unwrap();
}
