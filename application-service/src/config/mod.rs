pub mod server_config;
pub mod database_config;
pub mod app_config;
pub mod minio_config;

use std::env;
use server_config::ServerConfig;
use database_config::DatabaseConfig;
use app_config::AppConfig;
use minio_config::MinioConfig;

pub use server_config::*;
pub use database_config::*;
pub use app_config::*;
pub use minio_config::*;

/// Main application configuration combining all config modules
#[derive(Debug, Clone)]
pub struct Config {
    pub database_url: String,
    pub server_host: String,
    pub server_port: u16,
    pub grpc_port: u16,
    pub upload_dir: String,
    pub max_file_size: usize,
    pub cors_origins: Vec<String>,
    pub log_level: String,
    pub minio: MinioConfig,
}

impl Config {
    /// Load configuration from environment variables
    pub fn from_env() -> Result<Self, env::VarError> {
        dotenvy::dotenv().ok(); // Load .env file if present

        tracing::info!("Loading configuration from environment variables");

        let server_config = ServerConfig::from_env();
        let database_config = DatabaseConfig::from_env();
        let app_config = AppConfig::from_env();
        let minio_config = MinioConfig::from_env();

        let config = Config {
            database_url: database_config.url,
            server_host: server_config.host,
            server_port: server_config.port,
            grpc_port: server_config.grpc_port,
            upload_dir: app_config.upload_dir,
            max_file_size: app_config.max_file_size,
            cors_origins: app_config.cors_origins,
            log_level: app_config.log_level,
            minio: minio_config,
        };

        tracing::info!("Configuration loaded successfully");
        tracing::debug!("Config: {:?}", config);

        Ok(config)
    }
}
