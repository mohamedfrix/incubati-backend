use serde::{Deserialize, Serialize};
use crate::models::user::Role;
use uuid::Uuid;
use chrono::{DateTime, Utc};

#[derive(Debug, Deserialize)]
pub struct RegisterUserDto {
    pub email: String,
    pub nom: String,
    pub prenom: String,
    pub password: String,
    pub role: Role,
}

#[derive(Debug, Deserialize)]
pub struct LoginDto {
    pub email: String,
    pub password: String,
}

#[derive(Debug, Serialize)]
pub struct UserResponseDto {
    pub id: Uuid,
    pub email: String,
    pub nom: String,
    pub prenom: String,
    pub role: Role,
    pub created_by: Option<Uuid>,
    pub created_at: DateTime<Utc>,
} 