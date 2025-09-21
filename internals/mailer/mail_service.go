package mailer

import (
	"fmt"
	"log"
	"net/smtp"
	"time"
)

// The struct implements the Mail service for the application
type MailService struct {
	Host     string // e.g., "smtp.gmail.com"
	Port     int    // e.g., 587
	Username string
	Password string // The app password for the Gmail
	From     string // sender email address
	Expiry   time.Duration
}

// Reuest body for an email
type EmailRequest struct {
	To      string // Receiver address
	Subject string // Subject for the email
	Body    []byte // []byte gives the flexibility to use Txt/HTML template
}

// This creates a new Mail service for the login
func NewMailService(from, host, user, pass string, port int) *MailService {
	return &MailService{
		Host:     host,
		Port:     port,
		Username: user,
		Password: pass,
		From:     from,
		Expiry:   time.Second * 10,
	}
}

// Verifies the Mailer Login.
// Populate the Email with details.
// Push the mail through smtp server. Implement a retry mechanism in API layer
func (m *MailService) PushMail(req EmailRequest) error {
	// Set up authentication information.
	auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)
	// build the addr
	addr := fmt.Sprintf("%s:%d", m.Host, m.Port)
	// Connect to the server, authenticate, and send the email
	to := []string{req.To}
	err := smtp.SendMail(addr, auth, m.From, to, req.Body)
	if err != nil {
		log.Printf("error sending email: %v", err)
		return err
	}

	// email sent successfully
	return nil
}

// Invitation mail payload needs the token to be verified
// TODO: CHANGE THE VERIFICATION ROUTE
func InviteBody(token string) []byte {
	return []byte("From: CampaignHub Team <no-reply@campaignhub.com>\r\n" +
		"To: user@example.com\r\n" +
		"Subject: Verify your account\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		"<!DOCTYPE html>" +
		"<html>" +
		"<body style='margin:0;padding:0;background-color:#f4f4f7;font-family:Arial,sans-serif;'>" +
		"<table align='center' width='100%' cellpadding='0' cellspacing='0' style='max-width:600px;margin:20px auto;background:#ffffff;border-radius:8px;box-shadow:0 2px 4px rgba(0,0,0,0.1);'>" +
		"  <tr>" +
		"    <td style='padding:30px;text-align:center;border-bottom:1px solid #eaeaea;'>" +
		"      <h1 style='margin:0;font-size:24px;color:#333;'>Welcome to CampaignHub 🎉</h1>" +
		"    </td>" +
		"  </tr>" +
		"  <tr>" +
		"    <td style='padding:30px;color:#555;font-size:16px;line-height:1.5;'>" +
		"      <p>Hi there,</p>" +
		"      <p>Thanks for signing up! Please verify your account by clicking the button below:</p>" +
		"      <p style='text-align:center;margin:30px 0;'>" +
		"        <a href='" + fmt.Sprintf("http://localhost:8080/verify?token=%s", token) + "' " +
		"           style='display:inline-block;padding:14px 28px;font-size:16px;font-weight:bold;color:#ffffff;" +
		"           text-decoration:none;background-color:#4CAF50;border-radius:6px;'>Verify Account</a>" +
		"      </p>" +
		"      <p>If you did not sign up, you can safely ignore this email.</p>" +
		"      <p style='margin-top:40px;color:#999;font-size:14px;'>The CampaignHub Team</p>" +
		"    </td>" +
		"  </tr>" +
		"</table>" +
		fmt.Sprintf(
			`<p style='text-align:center;margin-top:20px;color:#aaa;font-size:12px;'>
			© %d CampaignHub. All rights reserved.</p>`,
			time.Now().Year(),
		) +
		"</body>" +
		"</html>")
}
