use lettre::{AsyncSmtpTransport, AsyncTransport, Message, Tokio1Executor, message::Mailbox};

/// Send an email using Gmail SMTP
pub async fn send_email(to: &str, subject: &str, body: &str, smtp_transport: &AsyncSmtpTransport<Tokio1Executor>, from: &str) -> Result<(), lettre::transport::smtp::Error> {
    let email = Message::builder()
        .from(from.parse::<Mailbox>().unwrap())
        .to(to.parse::<Mailbox>().unwrap())
        .subject(subject)
        .body(body.to_string())
        .unwrap();
    smtp_transport.send(email).await.map(|_| ())
}

// Note: Credentials and SMTP config should be loaded from environment/config. 