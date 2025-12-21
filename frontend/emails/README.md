# Email Templates Build

This directory contains React Email templates that are built into static HTML for use in the backend email service.

## Templates

- **OTPVerificationEmail.tsx**: Email sent to users with their 6-digit OTP code
- **EmailLayout.tsx**: Shared layout component for all emails

## Building Templates

To build the email templates to HTML:

```bash
cd frontend
npm run email:build
```

This will generate static HTML files in `emails/build/` that can be used by the backend email service.

## Development

To preview emails during development:

```bash
cd frontend
npm run email:dev
```

This will start the React Email development server at `http://localhost:3000`.

## Installation

If React Email dependencies are not installed:

```bash
cd frontend
npm install @react-email/components react-email
```

## Usage in Backend

The built HTML templates are used by the Brevo email service in:
`backend/infrastructure/services/brevo_email.go`

## Template Variables

### OTPVerificationEmail
- `otpCode` (string): The 6-digit verification code
- `organizationName` (string): Name of the organization being registered
- `recipientName` (string): Name of the recipient (admin user)
- `expirationMinutes` (number): Minutes until OTP expires (default: 10)

## Styling

Email templates use inline styles for maximum email client compatibility. The design follows:
- Clean, professional aesthetic
- Mobile-responsive layout
- High contrast for accessibility
- Security-focused messaging
