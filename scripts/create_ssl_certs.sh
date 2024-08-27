#!/bin/bash

# Set variables
CERT_DIR="src/certs"
PRIVATE_KEY="$CERT_DIR/privkey.pem"
CERTIFICATE="$CERT_DIR/fullchain.pem"
DAYS_VALID=365

# Create directory for certificates if it doesn't exist
mkdir -p $CERT_DIR

# Generate private key and certificate
openssl req -x509 -newkey rsa:4096 -nodes -keyout $PRIVATE_KEY -out $CERTIFICATE -days $DAYS_VALID -subj "/CN=localhost" -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

# Set appropriate permissions
chmod 644 $PRIVATE_KEY $CERTIFICATE

echo "SSL certificates created successfully:"
echo "Private Key: $PRIVATE_KEY"
echo "Certificate: $CERTIFICATE"
