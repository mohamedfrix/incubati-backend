mod config;
mod app;
mod models;
mod repository;
mod service;
mod dto;
mod routes;
mod utils;
mod middleware;
mod startup;
mod seeder;

use crate::config::Config;
use crate::startup::initialize_database;
use axum::serve;
use tokio::net::TcpListener;
use tracing_subscriber::fmt;
use tracing_subscriber::EnvFilter;
use tracing::{info, debug, warn, error};
use std::net::SocketAddr;
use sqlx::postgres::PgPoolOptions;
use std::sync::Arc;

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

    // Connect to database
    debug!("Connecting to database...");
    let pool = Arc::new(
        PgPoolOptions::new()
            .max_connections(5)
            .connect(&config.database_url)
            .await
            .expect("Failed to connect to database"),
    );
    info!("Database connection established");

    // Initialize database (migrations + admin user)
    match initialize_database(pool.clone()).await {
        Ok(()) => info!("Database initialization completed successfully"),
        Err(e) => {
            error!("Failed to initialize database: {}", e);
            std::process::exit(1);
        }
    }

    // Build the Axum app with all routes
    debug!("Building application routes...");
    let app = app::create_app(&config).await;
    info!("App and routes initialized");

    // Start the server
    let addr = SocketAddr::from(([0, 0, 0, 0], 4000));
    debug!(?addr, "Binding TCP listener");
    let listener = match TcpListener::bind(addr).await {
        Ok(listener) => listener,
        Err(e) => {
            error!("Failed to bind to address {}: {}", addr, e);
            return;
        }
    };
    info!("Auth Service running on {}", addr);
    if let Err(e) = serve(listener, app).await {
        error!("Server error: {}", e);
    }
}
