use serde::{Deserialize, Serialize};

/// Application status enumeration
#[derive(Debug, Clone, Serialize, Deserialize, sqlx::Type, PartialEq)]
#[sqlx(type_name = "application_status", rename_all = "snake_case")]
pub enum ApplicationStatus {
    Brouillon,           // Draft
    EnAttente,          // Pending review
    Approuvee,          // Approved
    Rejetee,            // Rejected
    ModificationDemandee, // Modification required
}

impl std::fmt::Display for ApplicationStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ApplicationStatus::Brouillon => write!(f, "Brouillon"),
            ApplicationStatus::EnAttente => write!(f, "En Attente"),
            ApplicationStatus::Approuvee => write!(f, "Approuvée"),
            ApplicationStatus::Rejetee => write!(f, "Rejetée"),
            ApplicationStatus::ModificationDemandee => write!(f, "Modification Demandée"),
        }
    }
}

/// Application type enumeration
#[derive(Debug, Clone, Serialize, Deserialize, sqlx::Type, PartialEq)]
#[sqlx(type_name = "application_type", rename_all = "PascalCase")]
pub enum ApplicationType {
    Internship,
    Incubation,
    Pfe,
}

impl ApplicationStatus {
    /// Convert to database string representation
    pub fn to_db_string(&self) -> &'static str {
        match self {
            ApplicationStatus::Brouillon => "brouillon",
            ApplicationStatus::EnAttente => "en_attente",
            ApplicationStatus::Approuvee => "approuvee",
            ApplicationStatus::Rejetee => "rejetee",
            ApplicationStatus::ModificationDemandee => "modification_demandee",
        }
    }
}

impl ApplicationType {
    /// Convert to database string representation
    pub fn to_db_string(&self) -> &'static str {
        match self {
            ApplicationType::Internship => "internship",
            ApplicationType::Incubation => "incubation",
            ApplicationType::Pfe => "pfe",
        }
    }
}

impl std::fmt::Display for ApplicationType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ApplicationType::Internship => write!(f, "Stage"),
            ApplicationType::Incubation => write!(f, "Incubation"),
            ApplicationType::Pfe => write!(f, "PFE"),
        }
    }
}
