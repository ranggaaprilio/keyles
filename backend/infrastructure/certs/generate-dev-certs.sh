#!/bin/bash
set -e

# Generate self-signed certificates for local development TLS testing
# Usage: ./generate-dev-certs.sh

CERT_DIR="$(cd "$(dirname "$0")/dev-certs" && pwd)"
KEY_FILE="$CERT_DIR/localhost.key"
CERT_FILE="$CERT_DIR/localhost.crt"

echo "Generating self-signed certificates for local development..."

openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout "$KEY_FILE" \
  -out "$CERT_FILE" \
  -subj "/C=US/ST=Dev/L=Local/O=Keyles/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

echo "✓ Certificates generated:"
echo "  - Private key: $KEY_FILE"
echo "  - Certificate: $CERT_FILE"
echo ""
echo "These certificates are for local development only."
echo "Do NOT commit them to version control."
