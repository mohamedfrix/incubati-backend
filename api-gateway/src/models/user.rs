use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use sqlx::FromRow;
use uuid::Uuid;

/// User roles in the system
#[derive(Debug, Clone, Serialize, Deserialize, sqlx::Type, PartialEq, Eq)]
#[sqlx(type_name = "user_role", rename_all = "lowercase")]
pub enum Role {
    Admin,
    Encadrant,
    Incube,
    Etudiant,
}

/// User model for authentication
#[derive(Debug, Clone, Serialize, Deserialize, FromRow)]
pub struct User {
    pub id: Uuid,
    pub email: String,
    pub nom: String,
    pub prenom: String,
    pub password: String, // hashed
    pub role: Role,
    pub created_at: DateTime<Utc>,
    pub created_by: Option<Uuid>,
} 