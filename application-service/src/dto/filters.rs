use serde::Deserialize;
use validator::Validate;

use crate::models::{ApplicationStatus, ApplicationType};

/// Query parameters for filtering applications
#[derive(Debug, Clone, Deserialize, Validate)]
pub struct ApplicationFilters {
    pub user_id: Option<String>,
    pub application_type: Option<ApplicationType>,
    pub status: Option<ApplicationStatus>,
    #[validate(range(min = 1))]
    pub page: Option<u32>,
    #[validate(range(min = 1, max = 100))]
    pub limit: Option<u32>,
}
