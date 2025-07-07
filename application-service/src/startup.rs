use sqlx::PgPool;
use tracing::{info, warn, error};

/// Initialize the database by running migrations
pub async fn initialize_database(pool: &PgPool) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    info!("Starting application service database initialization...");

    // Run migrations
    match run_migrations(pool).await {
        Ok(()) => info!("Application service database migrations completed successfully"),
        Err(e) => {
            error!("Failed to run application service migrations: {}", e);
            return Err(e);
        }
    }

    info!("Application service database initialization completed");
    Ok(())
}

/// Run database migrations with proper error handling
async fn run_migrations(pool: &PgPool) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    info!("Running application service database migrations...");
    
    match sqlx::migrate!("./migrations").run(pool).await {
        Ok(()) => {
            info!("All application service migrations applied successfully");
            Ok(())
        }
        Err(sqlx::migrate::MigrateError::VersionMissing(version)) => {
            warn!("Migration version {} is missing, but this might be expected", version);
            Ok(())
        }
        Err(e) => {
            // Check if the error is about migrations already being applied
            let error_msg = e.to_string();
            if error_msg.contains("already applied") || 
               error_msg.contains("no pending migrations") ||
               error_msg.contains("migration has already been applied") ||
               error_msg.contains("relation") && error_msg.contains("already exists") {
                info!("Application service migrations already applied, skipping");
                Ok(())
            } else {
                error!("Application service migration error: {}", e);
                Err(Box::new(e))
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use sqlx::postgres::PgPoolOptions;

    async fn create_test_pool() -> PgPool {
        let database_url = std::env::var("TEST_DATABASE_URL")
            .unwrap_or_else(|_| "postgresql://test:test@localhost:5432/test_app_service_db".to_string());
        
        PgPoolOptions::new()
            .max_connections(1)
            .connect(&database_url)
            .await
            .expect("Failed to connect to test database")
    }

    #[tokio::test]
    #[ignore] // Requires test database
    async fn test_initialize_database() {
        let pool = create_test_pool().await;
        let result = initialize_database(&pool).await;
        assert!(result.is_ok());
    }

    #[tokio::test]
    #[ignore] // Requires test database
    async fn test_run_migrations_twice() {
        let pool = create_test_pool().await;
        
        // Run migrations first time
        let result1 = run_migrations(&pool).await;
        assert!(result1.is_ok());
        
        // Run migrations second time - should not fail
        let result2 = run_migrations(&pool).await;
        assert!(result2.is_ok());
    }
}
