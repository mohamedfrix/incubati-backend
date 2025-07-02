// This file should be moved to src/bin/seeder.rs for correct crate root access.

// mod user;

use api_gateway::config::Config;
use api_gateway::seeder::user::seed_admin_user;
use sqlx::postgres::PgPoolOptions;
use std::sync::Arc;

#[tokio::main]
async fn main() {
    // Load config
    let config = Config::from_env();
    // Connect to DB
    let pool = Arc::new(
        PgPoolOptions::new()
            .max_connections(5)
            .connect(&config.database_url)
            .await
            .expect("Failed to connect to DB"),
    );
    // Seed admin user
    seed_admin_user(pool).await.expect("Failed to seed admin user");
    println!("Seeder finished.");
} 