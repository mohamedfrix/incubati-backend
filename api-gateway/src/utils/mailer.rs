use lettre::{AsyncSmtpTransport, AsyncTransport, Message, Tokio1Executor, message::Mailbox};

/// Send an email using Gmail SMTP
pub async fn send_email(to: &str, subject: &str, body: &str, smtp_transport: &AsyncSmtpTransport<Tokio1Executor>, from: &str) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let from_mailbox = from.parse::<Mailbox>()
        .map_err(|e| format!("Invalid from email address '{}': {}", from, e))?;
    let to_mailbox = to.parse::<Mailbox>()
        .map_err(|e| format!("Invalid to email address '{}': {}", to, e))?;
    
    let email = Message::builder()
        .from(from_mailbox)
        .to(to_mailbox)
        .subject(subject)
        .body(body.to_string())
        .map_err(|e| format!("Failed to build email message: {}", e))?;
    
    smtp_transport.send(email).await
        .map_err(|e| format!("Failed to send email: {}", e).into())
        .map(|_| ())
}

// Note: Credentials and SMTP config should be loaded from environment/config. 