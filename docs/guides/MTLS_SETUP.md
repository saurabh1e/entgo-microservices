# mTLS Configuration for Microservices
This guide explains how to enable mTLS (Mutual TLS) for secure service-to-service communication.
## Overview
mTLS provides:
- **Encryption**: All gRPC traffic is encrypted in transit
- **Authentication**: Both client and server verify each other's identity
- **Authorization**: Only services with valid certificates can communicate
- **Zero Trust**: No implicit trust based on network location
## Quick Start
### 1. Generate Certificates
```bash
cd /Users/saurabh/GolandProjects/entgo-microservices
./scripts/generate-certs.sh
```
This creates certificates for all services with proper CA signing.
### 2. Add to .gitignore
```bash
echo "certs/" >> .gitignore
```
**NEVER commit certificates to version control!**
### 3. Mount Certificates in Docker
Update your docker-compose.yml:
```yaml
services:
  auth:
    volumes:
      - ../certs/auth:/certs:ro
      - ../certs/ca:/certs/ca:ro
    environment:
      GRPC_TLS_ENABLED: "true"
      GRPC_SERVER_CERT: /certs/server-cert.pem
      GRPC_SERVER_KEY: /certs/server-key.pem
      GRPC_CA_CERT: /certs/ca/ca-cert.pem
```
## Security Best Practices
1. **Store CA private key securely** - Never commit to git
2. **Rotate certificates regularly** - Before 365 day expiry
3. **Monitor certificate expiry** - Set up alerts
4. **Use proper PKI in production** - Not self-signed certs
## Testing mTLS
```bash
# Verify certificate
openssl x509 -in certs/auth/server-cert.pem -text -noout
# Test with grpcurl
grpcurl \
  -cacert certs/ca/ca-cert.pem \
  -cert certs/gateway/client-cert.pem \
  -key certs/gateway/client-key.pem \
  localhost:9081 list
```
For full implementation details, see the main README.md.
