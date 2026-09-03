package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/resend/resend-go/v2"
	emailtemplates "github.com/trunglq04/goride/services/notification-service/email_templates"
	"github.com/trunglq04/goride/shared/env"
)

// EmailSender sends emails via the Resend API.
type EmailSender struct {
	client   *resend.Client
	fromAddr string
}

// NewEmailSender creates a new Resend-based email sender.
func NewEmailSender() *EmailSender {
	apiKey := env.GetString("RESEND_API_KEY", "", func(key string) {
		slog.Warn("RESEND_API_KEY not set, emails will not be sent", "key", key)
	})

	fromAddr := env.GetString("EMAIL_FROM_ADDRESS", "GoRide <noreply@goride.quangtrung.me>")

	client := resend.NewClient(apiKey)
	return &EmailSender{
		client:   client,
		fromAddr: fromAddr,
	}
}

// SendOTPEmail sends a 6-digit OTP verification email.
func (s *EmailSender) SendOTPEmail(ctx context.Context, toEmail, otp string) error {
	subject := "GoRide - Verify your email"
	htmlBody := fmt.Sprintf(emailtemplates.OtpSendingTemplate, otp)

	params := &resend.SendEmailRequest{
		From:    s.fromAddr,
		To:      []string{toEmail},
		Subject: subject,
		Html:    htmlBody,
	}

	sent, err := s.client.Emails.Send(params)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to send OTP email via Resend",
			"to", toEmail,
			"err", err,
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	slog.InfoContext(ctx, "OTP email sent successfully",
		"to", toEmail,
		"resend_id", sent.Id,
	)
	return nil
}
