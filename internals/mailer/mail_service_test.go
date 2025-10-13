package mailer

import (
	"testing"

	"github.com/Alter-Sitanshu/campaignHub/internals/db"
)

func TestPushMail(t *testing.T) {
	t.Run("testing mail service", func(t *testing.T) {
		mailer := NewMailService(
			"example@gmail.com",
			"localhost",
			"admin",
			"testing-pass",
			1025,
		)
		req := EmailRequest{
			To:      "test_subject@gmail.com",
			Subject: "Test_Subject",
			Body:    InviteBody("test_subject@gmail.com", "users", "example_token"),
		}
		err := mailer.PushMail(req)
		if err != nil {
			t.Fail()
		}
	})
	t.Run("testing invalid mail service", func(t *testing.T) {
		mailer := NewMailService(
			"",
			"localhost",
			"admin",
			"testing-pass",
			1025,
		)
		req := EmailRequest{
			To:      "",
			Subject: "Test_Subject",
			Body:    InviteBody("test_subject@gmail.com", "users", "example_token"),
		}
		err := mailer.PushMail(req)
		if err == nil {
			t.Fail()
		}
	})
}

func TestGenerateTicketEmail(t *testing.T) {
	t.Run("testing ticket email generation", func(t *testing.T) {
		mailer := NewMailService(
			"example@gmail.com",
			"localhost",
			"admin",
			"testing-pass",
			1025,
		)
		ticket := &db.Ticket{
			Id:        "TICKET123",
			Subject:   "Issue with login",
			Message:   "Unable to login with correct credentials.",
			CreatedAt: "2024-10-01 10:00:00",
			Type:      "user",
			Status:    1,
		}
		email := "example@gmail.com"
		body := GenerateTicketEmail(email, *ticket)
		if len(body) == 0 {
			t.Fail()
		}
		req := EmailRequest{
			To:      email,
			Subject: "New Ticket Raised - " + ticket.Id,
			Body:    body,
		}
		err := mailer.PushMail(req)
		if err != nil {
			t.Fail()
		}
	})
}

func TestResetPasswordBody(t *testing.T) {
	t.Run("testing reset password email generation", func(t *testing.T) {
		mailer := NewMailService(
			"example@gmail.com",
			"localhost",
			"admin",
			"testing-pass",
			1025,
		)

		email := "example@gmail.com"
		body := GeneratePasswordResetEmail(email, "random_token_123")
		if len(body) == 0 {
			t.Fail()
		}
		req := EmailRequest{
			To:      email,
			Subject: "Password Reset Request",
			Body:    body,
		}
		err := mailer.PushMail(req)
		if err != nil {
			t.Fail()
		}
	})
}
