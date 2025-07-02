use crate::models::user::{User, Role};
use sqlx::PgPool;
use uuid::Uuid;
use std::sync::Arc;
use tracing::{info, error, instrument};

pub struct UserRepository {
    pub pool: Arc<PgPool>,
}

impl UserRepository {
    pub fn new(pool: Arc<PgPool>) -> Self {
        Self { pool }
    }

    /// Create a new user
    #[instrument(skip(self, user))]
    pub async fn create_user(&self, user: &User) -> Result<User, sqlx::Error> {
        info!(email = %user.email, "Creating user in DB");
        let rec = match sqlx::query_as::<_, User>(
            r#"INSERT INTO users (id, email, nom, prenom, password, role, created_at, created_by)
               VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
               RETURNING *"#
        )
        .bind(user.id)
        .bind(&user.email)
        .bind(&user.nom)
        .bind(&user.prenom)
        .bind(&user.password)
        .bind(&user.role)
        .bind(user.created_at)
        .bind(user.created_by)
        .fetch_one(self.pool.as_ref())
        .await {
            Ok(rec) => rec,
            Err(e) => {
                error!(error = %e, email = %user.email, "DB error during user creation");
                return Err(e);
            }
        };
        info!(user_id = %rec.id, "User created in DB");
        Ok(rec)
    }

    /// Find a user by email
    #[instrument(skip(self, email))]
    pub async fn find_by_email(&self, email: &str) -> Result<Option<User>, sqlx::Error> {
        info!(email = %email, "Finding user by email");
        let result = sqlx::query_as::<_, User>(
            r#"SELECT * FROM users WHERE email = $1"#
        )
        .bind(email)
        .fetch_optional(self.pool.as_ref())
        .await;
        match &result {
            Ok(Some(user)) => info!(user_id = %user.id, "User found by email"),
            Ok(None) => info!(email = %email, "No user found for email"),
            Err(e) => error!(error = %e, email = %email, "DB error during find_by_email"),
        }
        result
    }

    /// List users (optionally by creator)
    #[instrument(skip(self, created_by))]
    pub async fn list_users(&self, created_by: Option<Uuid>) -> Result<Vec<User>, sqlx::Error> {
        info!(created_by = ?created_by, "Listing users from DB");
        let result: Result<Vec<User>, sqlx::Error> = match created_by {
            Some(creator) => {
                sqlx::query_as::<_, User>(
                    r#"SELECT * FROM users WHERE created_by = $1"#
                )
                .bind(creator)
                .fetch_all(self.pool.as_ref())
                .await
            },
            None => {
                sqlx::query_as::<_, User>(
                    r#"SELECT * FROM users"#
                )
                .fetch_all(self.pool.as_ref())
                .await
            }
        };
        match &result {
            Ok(users) => info!(count = users.len(), "Users listed from DB"),
            Err(e) => error!(error = %e, created_by = ?created_by, "DB error during list_users"),
        }
        result
    }
} 