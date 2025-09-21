package mailer

import (
	"testing"
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
			Body:    InviteBody("example_token"),
		}
		err := mailer.PushMail(req)
		if err != nil {
			t.Fail()
		}
	})
}
