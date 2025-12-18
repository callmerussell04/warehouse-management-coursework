package service_test

import (
	"errors"
	"io"
	"log/slog"
	"net/smtp"
	"testing"

	"warehouse-management-system/internal/service"

	"github.com/stretchr/testify/assert"
)

func TestSMTPNotificationService_SendEmail(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("Mock Mode (No Config)", func(t *testing.T) {
		t.Setenv("SMTP_HOST", "")
		t.Setenv("SMTP_USERNAME", "")

		svc := service.NewSMTPNotificationService(discardLogger)

		svc.SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
			t.Fatal("SendMail should not be called in mock mode")
			return nil
		}

		err := svc.SendEmail("test@test.com", "Subj", "Body")
		assert.NoError(t, err)
	})

	t.Run("Success Send", func(t *testing.T) {
		t.Setenv("SMTP_HOST", "smtp.example.com")
		t.Setenv("SMTP_PORT", "587")
		t.Setenv("SMTP_USERNAME", "user")
		t.Setenv("SMTP_PASSWORD", "pass")
		t.Setenv("SMTP_FROM", "no-reply@example.com")

		svc := service.NewSMTPNotificationService(discardLogger)

		var capturedAddr string
		var capturedMsg string
		var capturedTo []string

		svc.SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
			capturedAddr = addr
			capturedTo = to
			capturedMsg = string(msg)
			return nil
		}

		to := "client@domain.com"
		subject := "Welcome"
		body := "Hello World"

		err := svc.SendEmail(to, subject, body)

		assert.NoError(t, err)

		assert.Equal(t, "smtp.example.com:587", capturedAddr)
		assert.Equal(t, []string{to}, capturedTo)

		assert.Contains(t, capturedMsg, "To: client@domain.com")
		assert.Contains(t, capturedMsg, "Subject: Welcome")
		assert.Contains(t, capturedMsg, "MIME-Version: 1.0")
		assert.Contains(t, capturedMsg, "\r\n\r\nHello World\r\n")
	})

	t.Run("SMTP Error", func(t *testing.T) {
		t.Setenv("SMTP_HOST", "smtp.example.com")
		t.Setenv("SMTP_USERNAME", "user")

		svc := service.NewSMTPNotificationService(discardLogger)

		expectedErr := errors.New("connection timeout")
		svc.SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
			return expectedErr
		}

		err := svc.SendEmail("test@test.com", "Subj", "Body")

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}
