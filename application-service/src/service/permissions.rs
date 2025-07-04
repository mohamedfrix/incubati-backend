use std::sync::Arc;
use uuid::Uuid;
use tracing::{info, warn};
use validator::Validate;

use crate::models::{Application, Document, ApplicationStatus};
use crate::repository::{ApplicationRepository, DocumentRepository};
use crate::errors::PermissionError;

/// Permission types for applications
#[derive(Debug, Clone, Copy)]
pub enum ApplicationPermission {
    Read,
    Update,
    Delete,
    ChangeStatus,
}

/// Permission types for documents
#[derive(Debug, Clone, Copy)]
pub enum DocumentPermission {
    Read,
    Create,
    Delete,
}

/// Permission validation service
pub struct PermissionValidator {
    application_repository: Arc<dyn ApplicationRepository>,
    document_repository: Arc<dyn DocumentRepository>,
}

impl PermissionValidator {
    /// Create a new permission validator
    pub fn new(
        application_repository: Arc<dyn ApplicationRepository>,
        document_repository: Arc<dyn DocumentRepository>,
    ) -> Self {
        Self {
            application_repository,
            document_repository,
        }
    }

    /// Validate application permissions
    pub async fn validate_application_permission(
        &self,
        app_id: &str,
        requesting_user_id: &str,
        requesting_user_role: &str,
        required_permission: ApplicationPermission,
    ) -> Result<Application, PermissionError> {
        info!(
            user_id = requesting_user_id,
            user_role = requesting_user_role,
            app_id = app_id,
            permission = ?required_permission,
            "Validating application permission"
        );

        // Parse UUID
        let app_uuid = Uuid::parse_str(app_id)
            .map_err(|_| PermissionError::NotFound)?;

        // Get application from database
        let app = self.application_repository
            .get_application_by_id(app_uuid)
            .await
            .map_err(PermissionError::DatabaseError)?
            .ok_or(PermissionError::NotFound)?;

        // Validate user role
        self.validate_user_role(requesting_user_role)?;

        let is_owner = app.user_id == requesting_user_id;
        let is_admin = requesting_user_role == "Admin" || requesting_user_role == "Supervisor";

        let has_permission = match required_permission {
            ApplicationPermission::Read => {
                is_admin || is_owner
            },
            ApplicationPermission::Update => {
                let is_draft = app.status == ApplicationStatus::Brouillon;
                is_admin || (is_owner && is_draft)
            },
            ApplicationPermission::Delete => {
                let is_draft = app.status == ApplicationStatus::Brouillon;
                is_admin || (is_owner && is_draft)
            },
            ApplicationPermission::ChangeStatus => {
                is_admin
            }
        };

        if has_permission {
            info!(
                user_id = requesting_user_id,
                app_id = app_id,
                permission = ?required_permission,
                "Permission validation successful"
            );
            Ok(app)
        } else {
            warn!(
                user_id = requesting_user_id,
                user_role = requesting_user_role,
                app_id = app_id,
                permission = ?required_permission,
                is_owner = is_owner,
                is_admin = is_admin,
                "Permission validation failed"
            );
            Err(PermissionError::AccessDenied)
        }
    }

    /// Validate document permissions
    pub async fn validate_document_permission(
        &self,
        doc_id: &str,
        requesting_user_id: &str,
        requesting_user_role: &str,
        required_permission: DocumentPermission,
    ) -> Result<(Document, Application), PermissionError> {
        info!(
            user_id = requesting_user_id,
            user_role = requesting_user_role,
            doc_id = doc_id,
            permission = ?required_permission,
            "Validating document permission"
        );

        // Parse UUID
        let doc_uuid = Uuid::parse_str(doc_id)
            .map_err(|_| PermissionError::NotFound)?;

        // Get document from database
        let doc = self.document_repository
            .get_document_by_id(doc_uuid)
            .await
            .map_err(PermissionError::DatabaseError)?
            .ok_or(PermissionError::NotFound)?;

        // Get associated application
        let app = self.application_repository
            .get_application_by_id(doc.application_id)
            .await
            .map_err(PermissionError::DatabaseError)?
            .ok_or(PermissionError::NotFound)?;

        // Validate user role
        self.validate_user_role(requesting_user_role)?;

        let is_owner = app.user_id == requesting_user_id;
        let is_admin = requesting_user_role == "Admin" || requesting_user_role == "Supervisor";

        let has_permission = match required_permission {
            DocumentPermission::Read => {
                is_admin || is_owner
            },
            DocumentPermission::Create => {
                is_admin || is_owner
            },
            DocumentPermission::Delete => {
                is_admin || is_owner
            }
        };

        if has_permission {
            info!(
                user_id = requesting_user_id,
                doc_id = doc_id,
                permission = ?required_permission,
                "Document permission validation successful"
            );
            Ok((doc, app))
        } else {
            warn!(
                user_id = requesting_user_id,
                user_role = requesting_user_role,
                doc_id = doc_id,
                permission = ?required_permission,
                is_owner = is_owner,
                is_admin = is_admin,
                "Document permission validation failed"
            );
            Err(PermissionError::AccessDenied)
        }
    }

    /// Validate document creation permission (requires application access)
    pub async fn validate_document_creation_permission(
        &self,
        application_id: &str,
        requesting_user_id: &str,
        requesting_user_role: &str,
    ) -> Result<Application, PermissionError> {
        info!(
            user_id = requesting_user_id,
            user_role = requesting_user_role,
            application_id = application_id,
            "Validating document creation permission"
        );

        // Parse UUID
        let app_uuid = Uuid::parse_str(application_id)
            .map_err(|_| PermissionError::NotFound)?;

        // Get application from database
        let app = self.application_repository
            .get_application_by_id(app_uuid)
            .await
            .map_err(PermissionError::DatabaseError)?
            .ok_or(PermissionError::NotFound)?;

        // Validate user role
        self.validate_user_role(requesting_user_role)?;

        let is_owner = app.user_id == requesting_user_id;
        let is_admin = requesting_user_role == "Admin" || requesting_user_role == "Supervisor";

        if is_admin || is_owner {
            info!(
                user_id = requesting_user_id,
                application_id = application_id,
                "Document creation permission validation successful"
            );
            Ok(app)
        } else {
            warn!(
                user_id = requesting_user_id,
                user_role = requesting_user_role,
                application_id = application_id,
                is_owner = is_owner,
                is_admin = is_admin,
                "Document creation permission validation failed"
            );
            Err(PermissionError::AccessDenied)
        }
    }

    /// Validate user role
    fn validate_user_role(&self, role: &str) -> Result<(), PermissionError> {
        let valid_roles = ["Admin", "Supervisor", "Student", "User"];
        
        if valid_roles.contains(&role) {
            Ok(())
        } else {
            warn!("Invalid user role: {}", role);
            Err(PermissionError::InvalidRole { role: role.to_string() })
        }
    }
} 