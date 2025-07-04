use crate::models::{ApplicationType, ApplicationStatus};

/// Parse application type from database string
pub fn parse_application_type(type_str: &str) -> Option<ApplicationType> {
    match type_str {
        "internship" => Some(ApplicationType::Internship),
        "incubation" => Some(ApplicationType::Incubation),
        "pfe" => Some(ApplicationType::Pfe),
        _ => {
            tracing::warn!("Unknown application type: {}", type_str);
            None
        }
    }
}

/// Parse application status from database string
pub fn parse_application_status(status_str: &str) -> Option<ApplicationStatus> {
    match status_str {
        "brouillon" => Some(ApplicationStatus::Brouillon),
        "en_attente" => Some(ApplicationStatus::EnAttente),
        "approuvee" => Some(ApplicationStatus::Approuvee),
        "rejetee" => Some(ApplicationStatus::Rejetee),
        "modification_demandee" => Some(ApplicationStatus::ModificationDemandee),
        _ => {
            tracing::warn!("Unknown application status: {}", status_str);
            None
        }
    }
}
