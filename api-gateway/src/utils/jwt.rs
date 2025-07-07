use chrono::{Utc, Duration};
use jsonwebtoken::{encode, decode, Header, Validation, EncodingKey, DecodingKey, TokenData, errors::Result as JwtResult};
use serde::{Serialize, Deserialize};
use uuid::Uuid;

/// JWT Claims structure
#[derive(Debug, Serialize, Deserialize)]
pub struct Claims {
    pub sub: Uuid,
    pub email: String,
    pub role: String,
    pub exp: usize,
}

/// Create a JWT access token
pub fn create_access_token(user_id: Uuid, email: &str, role: &str, secret: &str, expires_in_minutes: i64) -> Result<String, jsonwebtoken::errors::Error> {
    let expiration = Utc::now() + Duration::minutes(expires_in_minutes);
    let claims = Claims {
        sub: user_id,
        email: email.to_owned(),
        role: role.to_owned(),
        exp: expiration.timestamp() as usize,
    };
    encode(&Header::default(), &claims, &EncodingKey::from_secret(secret.as_ref()))
}

/// Create a JWT refresh token
pub fn create_refresh_token(user_id: Uuid, email: &str, role: &str, secret: &str, expires_in_minutes: i64) -> Result<String, jsonwebtoken::errors::Error> {
    let expiration = Utc::now() + Duration::minutes(expires_in_minutes);
    let claims = Claims {
        sub: user_id,
        email: email.to_owned(),
        role: role.to_owned(),
        exp: expiration.timestamp() as usize,
    };
    encode(&Header::default(), &claims, &EncodingKey::from_secret(secret.as_ref()))
}

/// Validate a JWT and return the claims
pub fn validate_token(token: &str, secret: &str) -> JwtResult<TokenData<Claims>> {
    decode::<Claims>(token, &DecodingKey::from_secret(secret.as_ref()), &Validation::default())
}

/// Validate a refresh token and return the claims
pub fn validate_refresh_token(token: &str, secret: &str) -> JwtResult<TokenData<Claims>> {
    decode::<Claims>(token, &DecodingKey::from_secret(secret.as_ref()), &Validation::default())
}

pub use jsonwebtoken::errors; 