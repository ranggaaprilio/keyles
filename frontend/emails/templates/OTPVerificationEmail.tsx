import { Section, Text, Hr } from '@react-email/components';
import * as React from 'react';
import { EmailLayout } from '../components/EmailLayout';

interface OTPVerificationEmailProps {
  otpCode: string;
  organizationName: string;
  recipientName?: string;
  expirationMinutes?: number;
}

export const OTPVerificationEmail = ({
  otpCode = '123456',
  organizationName = 'Your Organization',
  recipientName = 'Admin',
  expirationMinutes = 10,
}: OTPVerificationEmailProps) => {
  return (
    <EmailLayout
      preview={`Your verification code is ${otpCode}`}
      heading="Verify Your Email Address"
    >
      <Text style={greeting}>Hello {recipientName},</Text>

      <Text style={paragraph}>
        Welcome to <strong>{organizationName}</strong>! To complete your registration and
        activate your organization's SSO platform, please use the verification code below:
      </Text>

      {/* OTP Code Display */}
      <Section style={otpContainer}>
        <Text style={otpCode}>{otpCode}</Text>
      </Section>

      <Text style={paragraph}>
        This code will expire in <strong>{expirationMinutes} minutes</strong>.
      </Text>

      <Hr style={divider} />

      <Text style={instructionTitle}>What to do next:</Text>
      <Text style={instruction}>
        1. Return to the verification page where you registered
      </Text>
      <Text style={instruction}>
        2. Enter the 6-digit code shown above
      </Text>
      <Text style={instruction}>
        3. Click "Verify" to activate your account
      </Text>

      <Hr style={divider} />

      <Section style={troubleshooting}>
        <Text style={troubleshootingTitle}>Didn't request this code?</Text>
        <Text style={paragraph}>
          If you didn't register for Keyles, you can safely ignore this email. The verification
          code will expire automatically.
        </Text>

        <Text style={troubleshootingTitle}>Need help?</Text>
        <Text style={paragraph}>
          If you're having trouble with verification, please contact our support team at{' '}
          <a href="mailto:support@keyles.io" style={link}>
            support@keyles.io
          </a>
          .
        </Text>
      </Section>

      <Hr style={divider} />

      <Text style={securityNote}>
        <strong>Security tip:</strong> Never share this code with anyone. Keyles will never ask
        you for your verification code via email, phone, or any other means.
      </Text>
    </EmailLayout>
  );
};

// Styles
const greeting = {
  color: '#1a1a1a',
  fontSize: '16px',
  fontWeight: '500',
  margin: '0 0 16px',
};

const paragraph = {
  color: '#374151',
  fontSize: '15px',
  lineHeight: '24px',
  margin: '0 0 16px',
};

const otpContainer = {
  backgroundColor: '#f3f4f6',
  borderRadius: '8px',
  padding: '24px',
  margin: '32px 0',
  textAlign: 'center' as const,
  border: '2px dashed #d1d5db',
};

const otpCode = {
  color: '#1a1a1a',
  fontSize: '40px',
  fontWeight: '700',
  letterSpacing: '8px',
  margin: '0',
  fontFamily: 'monospace',
};

const divider = {
  borderColor: '#e5e7eb',
  margin: '32px 0',
};

const instructionTitle = {
  color: '#1a1a1a',
  fontSize: '16px',
  fontWeight: '600',
  margin: '0 0 12px',
};

const instruction = {
  color: '#374151',
  fontSize: '14px',
  lineHeight: '22px',
  margin: '0 0 8px',
  paddingLeft: '8px',
};

const troubleshooting = {
  marginTop: '24px',
};

const troubleshootingTitle = {
  color: '#1a1a1a',
  fontSize: '15px',
  fontWeight: '600',
  margin: '0 0 8px',
};

const securityNote = {
  backgroundColor: '#fef3c7',
  borderLeft: '4px solid #f59e0b',
  padding: '12px 16px',
  borderRadius: '4px',
  color: '#78350f',
  fontSize: '13px',
  lineHeight: '20px',
  margin: '24px 0 0',
};

const link = {
  color: '#2563eb',
  textDecoration: 'underline',
};

export default OTPVerificationEmail;
