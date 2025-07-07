use sqlx::PgPool;
use std::sync::Arc;
use tracing::{info, warn, error};
use crate::seeder::user::seed_admin_user;

/// Initialize the database by running migrations and seeding initial data
pub async fn initialize_database(pool: Arc<PgPool>) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    info!("Starting database initialization...");

    // Run migrations
    match run_migrations(&pool).await {
        Ok(()) => info!("Database migrations completed successfully"),
        Err(e) => {
            error!("Failed to run migrations: {}", e);
            return Err(e);
        }
    }

    // Seed admin user
    match seed_admin_user(pool).await {
        Ok(()) => info!("Admin user seeding completed successfully"),
        Err(e) => {
            // This is not fatal - admin user might already exist
            warn!("Admin user seeding warning: {}", e);
        }
    }

    info!("Database initialization completed");
    Ok(())
}

/// Run database migrations with proper error handling
async fn run_migrations(pool: &PgPool) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    info!("Running database migrations...");
    
    match sqlx::migrate!("./migrations").run(pool).await {
        Ok(()) => {
            info!("All migrations applied successfully");
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
               error_msg.contains("migration has already been applied") {
                info!("Migrations already applied, skipping");
                Ok(())
            } else {
                error!("Migration error: {}", e);
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
            .unwrap_or_else(|_| "postgresql://test:test@localhost:5432/test_db".to_string());
        
        PgPoolOptions::new()
            .max_connections(1)
            .connect(&database_url)
            .await
            .expect("Failed to connect to test database")
    }

    #[tokio::test]
    #[ignore] // Requires test database
    async fn test_initialize_database() {
        let pool = Arc::new(create_test_pool().await);
        let result = initialize_database(pool).await;
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
