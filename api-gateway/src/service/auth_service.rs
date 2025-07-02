use crate::repository::user_repository::UserRepository;
use crate::dto::user::{RegisterUserDto, LoginDto, UserResponseDto};
use crate::dto::auth::AuthResponseDto;
use crate::models::user::{User, Role};
use crate::utils::{password, jwt};
use chrono::Utc;
use uuid::Uuid;
use sqlx::PgPool;
use std::sync::Arc;
use tracing::{info, error, debug, warn, instrument};

pub struct AuthService {
    pub user_repo: UserRepository,
    pub jwt_secret: String,
    pub access_token_expiry_minutes: i64,
    pub refresh_token_expiry_minutes: i64,
}

impl AuthService {
    pub fn new(pool: Arc<PgPool>, jwt_secret: String, access_token_expiry_minutes: i64, refresh_token_expiry_minutes: i64) -> Self {
        Self {
            user_repo: UserRepository::new(pool),
            jwt_secret,
            access_token_expiry_minutes,
            refresh_token_expiry_minutes,
        }
    }

    #[instrument(skip(self, dto, creator))]
    pub async fn register(&self, dto: RegisterUserDto, creator: &User) -> Result<AuthResponseDto, String> {
        debug!(email = %dto.email, role = ?dto.role, creator_id = %creator.id, "Service: register called");
        info!(email = %dto.email, role = ?dto.role, creator_id = %creator.id, "Registering new user");
        match creator.role {
            Role::Admin => {
                if dto.role == Role::Admin {
                    warn!(creator_id = %creator.id, "Admin attempted to create another admin");
                    error!("Admins cannot create other admins");
                    return Err("Admins cannot create other admins".into());
                }
            },
            Role::Encadrant => {
                if dto.role != Role::Etudiant && dto.role != Role::Incube {
                    warn!(creator_id = %creator.id, "Encadrant attempted to create invalid role: {:?}", dto.role);
                    error!("Encadrant can only create Etudiant or Incube");
                    return Err("Encadrant can only create Etudiant or Incube".into());
                }
            },
            _ => {
                warn!(creator_id = %creator.id, "User with insufficient permissions tried to register");
                error!("Insufficient permissions");
                return Err("Insufficient permissions".into())
            },
        }
        match self.user_repo.find_by_email(&dto.email).await {
            Ok(Some(_)) => {
                warn!(email = %dto.email, "Email already exists");
                error!("Email already exists");
                return Err("Email already exists".into());
            },
            Ok(None) => {},
            Err(e) => {
                error!(error = %e, "DB error while checking email existence");
                return Err("DB error".into());
            }
        }
        let hashed = password::hash_password(&dto.password).map_err(|e| {
            error!(error = ?e, "Password hash error");
            "Hash error"
        })?;
        debug!("Password hashed successfully");
        let user = User {
            id: Uuid::new_v4(),
            email: dto.email.clone(),
            nom: dto.nom.clone(),
            prenom: dto.prenom.clone(),
            password: hashed,
            role: dto.role.clone(),
            created_at: Utc::now(),
            created_by: Some(creator.id),
        };
        let user = match self.user_repo.create_user(&user).await {
            Ok(u) => u,
            Err(e) => {
                error!(error = %e, "DB error while creating user");
                return Err(e.to_string());
            }
        };
        debug!(user_id = %user.id, "User created in DB");
        let access_token = jwt::create_access_token(user.id, &user.email, &format!("{:?}", user.role).to_lowercase(), &self.jwt_secret, self.access_token_expiry_minutes);
        let refresh_token = jwt::create_refresh_token(user.id, &user.email, &format!("{:?}", user.role).to_lowercase(), &self.jwt_secret, self.refresh_token_expiry_minutes);
        info!(user_id = %user.id, "User registered successfully");
        Ok(AuthResponseDto {
            access_token,
            refresh_token,
            user: UserResponseDto {
                id: user.id,
                email: user.email,
                nom: user.nom,
                prenom: user.prenom,
                role: user.role,
                created_by: user.created_by,
                created_at: user.created_at,
            },
        })
    }

    #[instrument(skip(self, dto))]
    pub async fn login(&self, dto: LoginDto) -> Result<AuthResponseDto, String> {
        debug!(email = %dto.email, "Service: login called");
        info!(email = %dto.email, "Attempting login");
        let user = match self.user_repo.find_by_email(&dto.email).await {
            Ok(Some(u)) => u,
            Ok(None) => {
                warn!(email = %dto.email, "Invalid credentials: user not found");
                debug!("Returning invalid credentials error");
                return Err("Invalid credentials".into());
            },
            Err(e) => {
                error!(error = %e, email = %dto.email, "DB error during login");
                return Err("DB error".into());
            }
        };
        match password::verify_password(&user.password, &dto.password) {
            Ok(true) => debug!(email = %dto.email, "Password verified successfully"),
            Ok(false) => {
                warn!(email = %dto.email, "Invalid credentials: password mismatch");
                debug!("Returning invalid credentials error");
                return Err("Invalid credentials".into());
            },
            Err(e) => {
                error!(error = ?e, email = %dto.email, "Password verify error");
                return Err("Verify error".into());
            }
        }
        let access_token = jwt::create_access_token(user.id, &user.email, &format!("{:?}", user.role).to_lowercase(), &self.jwt_secret, self.access_token_expiry_minutes);
        let refresh_token = jwt::create_refresh_token(user.id, &user.email, &format!("{:?}", user.role).to_lowercase(), &self.jwt_secret, self.refresh_token_expiry_minutes);
        info!(user_id = %user.id, "Login successful");
        Ok(AuthResponseDto {
            access_token,
            refresh_token,
            user: UserResponseDto {
                id: user.id,
                email: user.email,
                nom: user.nom,
                prenom: user.prenom,
                role: user.role,
                created_by: user.created_by,
                created_at: user.created_at,
            },
        })
    }

    #[instrument(skip(self, token))]
    pub async fn verify_token(&self, token: &str) -> Result<UserResponseDto, String> {
        use jsonwebtoken::errors::ErrorKind;
        debug!("Service: verify_token called");
        info!("Verifying token");
        let claims = match jwt::validate_token(token, &self.jwt_secret) {
            Ok(data) => data.claims,
            Err(e) => {
                let dev_msg = match e.kind() {
                    ErrorKind::ExpiredSignature => "Token expired".to_string(),
                    ErrorKind::InvalidToken => "Malformed token".to_string(),
                    ErrorKind::InvalidSignature => "Invalid signature".to_string(),
                    ErrorKind::InvalidIssuer => "Invalid issuer".to_string(),
                    ErrorKind::InvalidAudience => "Invalid audience".to_string(),
                    _ => format!("Other JWT error: {}", e.to_string()),
                };
                warn!(error = ?e, "Token verification failed");
                error!(error = ?e, "Invalid token");
                return Err(dev_msg);
            }
        };
        debug!(email = %claims.email, "Token claims extracted");
        let user = match self.user_repo.find_by_email(&claims.email).await {
            Ok(Some(u)) => u,
            Ok(None) => {
                warn!(email = %claims.email, "User not found for token");
                error!("User not found for token");
                return Err("User not found for token".into());
            },
            Err(e) => {
                error!(error = %e, email = %claims.email, "DB error during token verification");
                return Err("DB error during token verification".into());
            }
        };
        info!(user_id = %user.id, "Token verified");
        Ok(UserResponseDto {
            id: user.id,
            email: user.email,
            nom: user.nom,
            prenom: user.prenom,
            role: user.role,
            created_by: user.created_by,
            created_at: user.created_at,
        })
    }

    #[instrument(skip(self, requester))]
    pub async fn list_users(&self, requester: &User) -> Result<Vec<UserResponseDto>, String> {
        debug!(requester_id = %requester.id, role = ?requester.role, "Service: list_users called");
        info!(requester_id = %requester.id, role = ?requester.role, "Listing users");
        let users = match requester.role {
            Role::Admin => self.user_repo.list_users(None).await,
            Role::Encadrant => self.user_repo.list_users(Some(requester.id)).await,
            _ => {
                warn!(requester_id = %requester.id, "Access denied for listing users");
                error!("Access denied");
                return Err("Access denied".into());
            },
        };
        let users = match users {
            Ok(u) => u,
            Err(e) => {
                error!(error = %e, requester_id = %requester.id, "DB error during list_users");
                return Err("DB error".into());
            }
        };
        info!(count = users.len(), "Users listed");
        Ok(users.into_iter().map(|user| UserResponseDto {
            id: user.id,
            email: user.email,
            nom: user.nom,
            prenom: user.prenom,
            role: user.role,
            created_by: user.created_by,
            created_at: user.created_at,
        }).collect())
    }

    #[instrument(skip(self, refresh_token))]
    pub async fn refresh_token(&self, refresh_token: String) -> Result<AuthResponseDto, String> {
        use jsonwebtoken::errors::ErrorKind;
        use crate::utils::jwt::validate_refresh_token;
        debug!("Service: refresh_token called");
        let claims = match validate_refresh_token(&refresh_token, &self.jwt_secret) {
            Ok(data) => data.claims,
            Err(e) => {
                let dev_msg = match e.kind() {
                    ErrorKind::ExpiredSignature => "Refresh token expired".to_string(),
                    ErrorKind::InvalidToken => "Malformed refresh token".to_string(),
                    ErrorKind::InvalidSignature => "Invalid refresh token signature".to_string(),
                    _ => format!("Other JWT error: {}", e.to_string()),
                };
                warn!(error = ?e, "Refresh token verification failed");
                debug!("Returning refresh token error");
                return Err(dev_msg);
            }
        };
        debug!(email = %claims.email, "Refresh token claims extracted");
        let user = match self.user_repo.find_by_email(&claims.email).await {
            Ok(Some(u)) => u,
            Ok(None) => {
                warn!(email = %claims.email, "User not found for refresh token");
                debug!("Returning user not found for refresh token error");
                return Err("User not found for refresh token".to_string());
            },
            Err(e) => {
                error!(error = %e, email = %claims.email, "DB error during refresh_token");
                return Err("DB error during refresh_token".to_string());
            }
        };
        let access_token = jwt::create_access_token(user.id, &user.email, &format!("{:?}", user.role).to_lowercase(), &self.jwt_secret, self.access_token_expiry_minutes);
        let refresh_token = jwt::create_refresh_token(user.id, &user.email, &format!("{:?}", user.role).to_lowercase(), &self.jwt_secret, self.refresh_token_expiry_minutes);
        info!(user_id = %user.id, "Refresh token successful");
        Ok(AuthResponseDto {
            access_token,
            refresh_token,
            user: UserResponseDto {
                id: user.id,
                email: user.email,
                nom: user.nom,
                prenom: user.prenom,
                role: user.role,
                created_by: user.created_by,
                created_at: user.created_at,
            },
        })
    }
} 