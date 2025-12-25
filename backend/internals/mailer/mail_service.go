package mailer

import (
	"fmt"
	"log"
	"time"

	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/resend/resend-go/v2"
)

// The struct implements the Mail service for the application
type MailService struct {
	client *resend.Client
}

// Reuest body for an email
type EmailRequest struct {
	To      string // Receiver address
	Subject string // Subject for the email
	Body    []byte // []byte gives the flexibility to use Txt/HTML template
}

// This creates a new Mail service for the login
func NewMailService(apiKey string) *MailService {
	return &MailService{
		client: resend.NewClient(apiKey),
	}
}

// Verifies the Mailer Login.
// Populate the Email with details.
// Push the mail through smtp server. Implement a retry mechanism in API layer
func (m *MailService) PushMail(req EmailRequest) error {
	// Connect to the server, authenticate, and send the email
	if req.To == "" {
		return fmt.Errorf("invalid address")
	}
	params := &resend.SendEmailRequest{
		From:    "FrogMedia <onboarding@resend.dev>",
		To:      []string{req.To},
		Subject: req.Subject,
		Html:    string(req.Body),
	}

	_, err := m.client.Emails.Send(params)
	if err != nil {
		log.Printf("mail sending failed: %v\n", err.Error())
		return err
	}

	// email sent successfully
	return nil
}

// Invitation mail payload needs the token to be verified
// TODO: CHANGE THE VERIFICATION ROUTE and email of user
func InviteBody(email, token string) []byte {
	return []byte(
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
			"        <a href='" + fmt.Sprintf("http://localhost:5173/auth/verify?token=%s", token) + "' " +
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

// Function generates the email template for the raised ticket notification
func GenerateTicketEmail(email string, ticket db.Ticket) []byte {
	return []byte(
		"<html>" +
			"<body style='font-family: Arial, sans-serif; background-color:#f9fafb; padding:20px;'>" +
			"<div style='max-width:600px; margin:auto; background:#ffffff; padding:20px; border-radius:8px; border:1px solid #e5e7eb;'>" +
			"<h2 style='color:#111827;'>🎫 New Ticket Raised</h2>" +

			"<p><strong>Ticket ID:</strong> " + ticket.Id + "</p>" +
			"<p><strong>Subject:</strong> " + ticket.Subject + "</p>" +
			"<p><strong>Reporter Email:</strong> " + email + "</p>" +
			"<p><strong>Message:</strong><br/>" + ticket.Message + "</p>" +
			"<p><strong>Created At:</strong> " + ticket.CreatedAt + "</p>" +
			"<p><strong>User Type:</strong> " + ticket.Type + "</p>" +

			"<p style='margin-top:20px; font-size:12px; color:#6b7280;'>This is an automated message from CampaignHub.</p>" +
			"</div>" +
			"</body>" +
			"</html>",
	)
}

// Funtion generates the email template for the password reset
// TODO: Change the reset link
func GeneratePasswordResetEmail(email, token string) []byte {
	resetLink := "http://localhost:5173/reset-password?token=" + token

	return []byte(
		"<html>" +
			"<body style='font-family: Arial, sans-serif; background-color:#f9fafb; padding:20px;'>" +
			"<div style='max-width:600px; margin:auto; background:#ffffff; padding:20px; border-radius:8px; border:1px solid #e5e7eb;'>" +
			"<h2 style='color:#111827;'>🔐 Password Reset Request</h2>" +

			"<p>Hello,</p>" +
			"<p>We received a request to reset your password. If you made this request, click the button below:</p>" +

			"<p><a href='" + resetLink + "' style='display:inline-block;padding:10px 20px;font-size:16px;color:#fff;" +
			"text-decoration:none;background-color:#4CAF50;border-radius:5px;'>Reset Password</a></p>" +

			"<p>If you didn’t request a password reset, you can safely ignore this email.</p>" +
			"<p style='margin-top:20px; font-size:12px; color:#6b7280;'>This link will expire in 15 minutes for your security.</p>" +
			"</div>" +
			"</body>" +
			"</html>",
	)
}
