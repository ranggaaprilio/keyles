import {
  Body,
  Container,
  Head,
  Heading,
  Html,
  Img,
  Link,
  Preview,
  Section,
  Text,
} from '@react-email/components';
import * as React from 'react';

interface EmailLayoutProps {
  preview: string;
  heading: string;
  children: React.ReactNode;
}

export const EmailLayout = ({ preview, heading, children }: EmailLayoutProps) => {
  return (
    <Html>
      <Head />
      <Preview>{preview}</Preview>
      <Body style={main}>
        <Container style={container}>
          {/* Header with Logo */}
          <Section style={header}>
            <Heading style={h1}>Keyles</Heading>
            <Text style={tagline}>Multi-Tenant SSO Platform</Text>
          </Section>

          {/* Main Content */}
          <Section style={content}>
            <Heading as="h2" style={h2}>
              {heading}
            </Heading>
            {children}
          </Section>

          {/* Footer */}
          <Section style={footer}>
            <Text style={footerText}>
              This email was sent from Keyles Multi-Tenant SSO Platform.
            </Text>
            <Text style={footerText}>
              If you didn't request this email, please ignore it or{' '}
              <Link href="mailto:support@keyles.io" style={link}>
                contact support
              </Link>
              .
            </Text>
            <Text style={footerCopyright}>
              © {new Date().getFullYear()} Keyles. All rights reserved.
            </Text>
          </Section>
        </Container>
      </Body>
    </Html>
  );
};

// Styles
const main = {
  backgroundColor: '#f6f9fc',
  fontFamily:
    '-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Ubuntu,sans-serif',
};

const container = {
  backgroundColor: '#ffffff',
  margin: '0 auto',
  padding: '20px 0 48px',
  marginBottom: '64px',
  maxWidth: '600px',
};

const header = {
  padding: '32px 40px',
  textAlign: 'center' as const,
  borderBottom: '1px solid #e6ebf1',
};

const h1 = {
  color: '#1a1a1a',
  fontSize: '32px',
  fontWeight: '700',
  margin: '0 0 8px',
  padding: '0',
  lineHeight: '1.2',
};

const tagline = {
  color: '#6b7280',
  fontSize: '14px',
  margin: '0',
  padding: '0',
};

const content = {
  padding: '40px 40px',
};

const h2 = {
  color: '#1a1a1a',
  fontSize: '24px',
  fontWeight: '600',
  margin: '0 0 24px',
  padding: '0',
  lineHeight: '1.3',
};

const footer = {
  padding: '24px 40px',
  borderTop: '1px solid #e6ebf1',
  textAlign: 'center' as const,
};

const footerText = {
  color: '#6b7280',
  fontSize: '12px',
  lineHeight: '20px',
  margin: '8px 0',
};

const footerCopyright = {
  color: '#9ca3af',
  fontSize: '11px',
  margin: '16px 0 0',
};

const link = {
  color: '#2563eb',
  textDecoration: 'underline',
};

export default EmailLayout;
