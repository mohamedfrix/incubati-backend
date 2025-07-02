use crate::models::user::{User, Role};
use crate::utils::password::hash_password;
use sqlx::PgPool;
use uuid::Uuid;
use chrono::Utc;
use std::sync::Arc;

pub async fn seed_admin_user(pool: Arc<PgPool>) -> Result<(), Box<dyn std::error::Error>> {
    let email = std::env::var("ADMIN_EMAIL").unwrap_or_else(|_| "admin@example.com".to_string());
    let password = std::env::var("ADMIN_PASSWORD").unwrap_or_else(|_| "adminpassword".to_string());
    let nom = std::env::var("ADMIN_NOM").unwrap_or_else(|_| "Admin".to_string());
    let prenom = std::env::var("ADMIN_PRENOM").unwrap_or_else(|_| "Admin".to_string());
    let hashed = hash_password(&password).map_err(|e| e.to_string())?;
    let user = User {
        id: Uuid::new_v4(),
        email,
        nom,
        prenom,
        password: hashed,
        role: Role::Admin,
        created_at: Utc::now(),
        created_by: None,
    };
    // Insert only if not exists
    let exists: Option<(Uuid,)> = sqlx::query_as("SELECT id FROM users WHERE email = $1")
        .bind(&user.email)
        .fetch_optional(pool.as_ref())
        .await?;
    if exists.is_none() {
        sqlx::query("INSERT INTO users (id, email, nom, prenom, password, role, created_at, created_by) VALUES ($1, $2, $3, $4, $5, $6::user_role, $7, $8)")
            .bind(user.id)
            .bind(&user.email)
            .bind(&user.nom)
            .bind(&user.prenom)
            .bind(&user.password)
            .bind("admin")
            .bind(user.created_at)
            .bind(user.created_by)
            .execute(pool.as_ref())
            .await?;
        println!("Seeded admin user: {}", user.email);
    } else {
        println!("Admin user already exists: {}", user.email);
    }
    Ok(())
} 