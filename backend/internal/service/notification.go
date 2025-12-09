package service

import (
	"log/slog"
	"net/smtp"
	"os"
)

type SMTPNotificationService struct {
	host     string
	port     string
	username string
	password string
	from     string
	logger   *slog.Logger
}

func NewSMTPNotificationService(logger *slog.Logger) *SMTPNotificationService {
	return &SMTPNotificationService{
		host:     os.Getenv("SMTP_HOST"),
		port:     os.Getenv("SMTP_PORT"),
		username: os.Getenv("SMTP_USERNAME"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     os.Getenv("SMTP_FROM"),
		logger:   logger,
	}
}

func (s *SMTPNotificationService) SendEmail(to, subject, body string) error {
	if s.host == "" || s.username == "" {
		s.logger.Info("[MOCK EMAIL] SMTP not configured", "to", to, "subject", subject, "body", body)
		return nil
	}

	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	address := s.host + ":" + s.port

	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		body + "\r\n")

	err := smtp.SendMail(address, auth, s.from, []string{to}, msg)
	if err != nil {
		s.logger.Error("failed to send email", "error", err, "to", to)
		return err
	}

	s.logger.Info("email sent successfully", "to", to)
	return nil
}
