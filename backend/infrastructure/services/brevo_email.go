package services

import (
	"context"
	"fmt"

	brevo "github.com/getbrevo/brevo-go/lib"
	"github.com/ranggaaprilio/keyles/domain/services"
)

// BrevoEmailService implements EmailService using Brevo (SendinBlue)
type BrevoEmailService struct {
	client      *brevo.APIClient
	senderEmail string
	senderName  string
}

// NewBrevoEmailService creates a new Brevo email service
func NewBrevoEmailService(apiKey, senderEmail, senderName string) services.EmailService {
	cfg := brevo.NewConfiguration()
	cfg.AddDefaultHeader("api-key", apiKey)
	
	return &BrevoEmailService{
		client:      brevo.NewAPIClient(cfg),
		senderEmail: senderEmail,
		senderName:  senderName,
	}
}

// SendOTPEmail sends an OTP verification email
func (s *BrevoEmailService) SendOTPEmail(ctx context.Context, toEmail, toName, otpCode, organizationName string) error {
	subject := fmt.Sprintf("Verify your email for %s", organizationName)
	
	htmlContent := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
		</head>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
			<div style="background-color: #f8f9fa; padding: 30px; border-radius: 10px;">
				<h1 style="color: #2563eb; margin-bottom: 20px;">Welcome to Keyles SSO!</h1>
				
				<p>Hello %s,</p>
				
				<p>Thank you for registering <strong>%s</strong> with Keyles SSO platform.</p>
				
				<p>To complete your registration and activate your tenant account, please use the following verification code:</p>
				
				<div style="background-color: #fff; border: 2px solid #2563eb; border-radius: 8px; padding: 20px; text-align: center; margin: 30px 0;">
					<h2 style="color: #2563eb; font-size: 36px; margin: 0; letter-spacing: 8px; font-family: 'Courier New', monospace;">%s</h2>
				</div>
				
				<p><strong>Important:</strong> This code will expire in <strong>10 minutes</strong>.</p>
				
				<p>If you didn't request this verification, please ignore this email or contact our support team.</p>
				
				<hr style="border: none; border-top: 1px solid #dee2e6; margin: 30px 0;">
				
				<p style="font-size: 12px; color: #6c757d;">
					Need help? Contact us at <a href="mailto:support@keyles.com" style="color: #2563eb;">support@keyles.com</a>
				</p>
			</div>
		</body>
		</html>
	`, toName, organizationName, otpCode)

	email := brevo.SendSmtpEmail{
		Sender: &brevo.SendSmtpEmailSender{
			Name:  s.senderName,
			Email: s.senderEmail,
		},
		To: []brevo.SendSmtpEmailTo{
			{
				Email: toEmail,
				Name:  toName,
			},
		},
		Subject:     subject,
		HtmlContent: htmlContent,
	}

	_, _, err := s.client.TransactionalEmailsApi.SendTransacEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	return nil
}

// SendWelcomeEmail sends a welcome email after successful verification
func (s *BrevoEmailService) SendWelcomeEmail(ctx context.Context, toEmail, toName, organizationName string) error {
	subject := fmt.Sprintf("Welcome to Keyles SSO - %s Activated!", organizationName)
	
	htmlContent := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
		</head>
		<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
			<div style="background-color: #f8f9fa; padding: 30px; border-radius: 10px;">
				<h1 style="color: #10b981; margin-bottom: 20px;">🎉 Your Tenant is Now Active!</h1>
				
				<p>Hello %s,</p>
				
				<p>Congratulations! Your organization <strong>%s</strong> has been successfully verified and activated on Keyles SSO platform.</p>
				
				<p>You can now:</p>
				<ul>
					<li>Log in to your tenant dashboard</li>
					<li>Configure SSO settings</li>
					<li>Manage your organization's identity provider</li>
					<li>Add team members</li>
				</ul>
				
				<div style="text-align: center; margin: 30px 0;">
					<a href="https://keyles.com/login" style="background-color: #2563eb; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Go to Dashboard</a>
				</div>
				
				<hr style="border: none; border-top: 1px solid #dee2e6; margin: 30px 0;">
				
				<p style="font-size: 12px; color: #6c757d;">
					Need help getting started? Contact us at <a href="mailto:support@keyles.com" style="color: #2563eb;">support@keyles.com</a>
				</p>
			</div>
		</body>
		</html>
	`, toName, organizationName)

	email := brevo.SendSmtpEmail{
		Sender: &brevo.SendSmtpEmailSender{
			Name:  s.senderName,
			Email: s.senderEmail,
		},
		To: []brevo.SendSmtpEmailTo{
			{
				Email: toEmail,
				Name:  toName,
			},
		},
		Subject:     subject,
		HtmlContent: htmlContent,
	}

	_, _, err := s.client.TransactionalEmailsApi.SendTransacEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	return nil
}
