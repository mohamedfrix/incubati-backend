use serde::Serialize;
use super::user::UserResponseDto;

#[derive(Debug, Serialize)]
pub struct AuthResponseDto {
    pub access_token: String,
    pub refresh_token: String,
    pub user: UserResponseDto,
} 