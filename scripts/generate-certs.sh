#!/bin/bash
# Generate mTLS certificates for microservices
# This script creates a Certificate Authority and service certificates for secure gRPC communication

set -e

CERTS_DIR="./certs"
DAYS_VALID=365

echo "🔐 Generating mTLS Certificates for Microservices"
echo "=================================================="

# Create directories
echo "📁 Creating certificate directories..."
mkdir -p $CERTS_DIR/ca
mkdir -p $CERTS_DIR/auth
mkdir -p $CERTS_DIR/gateway
mkdir -p $CERTS_DIR/orders

# Generate Certificate Authority
echo ""
echo "🏛️  Generating Certificate Authority (CA)..."
openssl genrsa -out $CERTS_DIR/ca/ca-key.pem 4096

openssl req -new -x509 -days $DAYS_VALID -key $CERTS_DIR/ca/ca-key.pem \
    -out $CERTS_DIR/ca/ca-cert.pem \
    -subj "/CN=Bolt Microservices CA/O=Bolt/C=US" \
    -addext "subjectAltName=DNS:ca.bolt.local"

echo "✅ CA certificate generated"

# Function to generate service certificates
generate_service_cert() {
    local service=$1
    local dir=$CERTS_DIR/$service

    echo ""
    echo "🔑 Generating certificates for $service service..."

    # Server certificate
    openssl genrsa -out $dir/server-key.pem 2048

    openssl req -new -key $dir/server-key.pem -out $dir/server.csr \
        -subj "/CN=$service/O=Bolt/C=US"

    # Create extension file for SAN
    cat > $dir/server-ext.cnf << EOF
subjectAltName = DNS:$service,DNS:entgo_${service}_dev,DNS:localhost,IP:127.0.0.1
extendedKeyUsage = serverAuth
EOF

    openssl x509 -req -days $DAYS_VALID -in $dir/server.csr \
        -CA $CERTS_DIR/ca/ca-cert.pem -CAkey $CERTS_DIR/ca/ca-key.pem \
        -CAcreateserial -out $dir/server-cert.pem \
        -extfile $dir/server-ext.cnf

    # Client certificate
    openssl genrsa -out $dir/client-key.pem 2048

    openssl req -new -key $dir/client-key.pem -out $dir/client.csr \
        -subj "/CN=$service-client/O=Bolt/C=US"

    # Create extension file for client
    cat > $dir/client-ext.cnf << EOF
extendedKeyUsage = clientAuth
EOF

    openssl x509 -req -days $DAYS_VALID -in $dir/client.csr \
        -CA $CERTS_DIR/ca/ca-cert.pem -CAkey $CERTS_DIR/ca/ca-key.pem \
        -CAcreateserial -out $dir/client-cert.pem \
        -extfile $dir/client-ext.cnf

    # Cleanup temporary files
    rm $dir/*.csr $dir/*.cnf

    echo "✅ Certificates for $service generated"
}

# Generate certificates for each service
generate_service_cert "auth"
generate_service_cert "gateway"
generate_service_cert "orders"

echo ""
echo "=================================================="
echo "✅ Certificate generation complete!"
echo ""
echo "📁 Certificates location: $CERTS_DIR/"
echo ""
echo "⚠️  IMPORTANT SECURITY NOTES:"
echo "   1. Keep $CERTS_DIR/ca/ca-key.pem SECURE"
echo "   2. DO NOT commit certificates to git"
echo "   3. Add $CERTS_DIR/ to .gitignore"
echo "   4. Rotate certificates before expiry ($DAYS_VALID days)"
echo ""
echo "📝 Next steps:"
echo "   1. Add $CERTS_DIR/ to .gitignore"
echo "   2. Mount certificates in docker-compose.yml"
echo "   3. Configure services to use mTLS"
echo "   4. Refer to README.md for usage instructions"
echo ""

