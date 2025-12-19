#!/bin/bash
# Login to auth service and print user details and tokens

set -e

# Configuration
GATEWAY_URL="http://localhost:8080/graphql"
EMAIL="${1:-admin@example.com}"
PASSWORD="${2:-admin123}"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}🔐 Logging in...${NC}"
echo "Email: $EMAIL"
echo ""

# GraphQL mutation
QUERY='mutation Login($email: String!, $password: String!) {
  login(input: { email: $email, password: $password }) {
    user {
      id
      firstName
      lastName
      email
      role
      isActive
      emailConfirmed
    }
    accessToken
    refreshToken
  }
}'

# Make the request
RESPONSE=$(curl -s -X POST "$GATEWAY_URL" \
  -H "Content-Type: application/json" \
  -d "{
    \"query\": $(echo "$QUERY" | jq -Rs .),
    \"variables\": {
      \"email\": \"$EMAIL\",
      \"password\": \"$PASSWORD\"
    }
  }")

# Check for errors
if echo "$RESPONSE" | jq -e '.errors' > /dev/null 2>&1; then
  echo -e "${YELLOW}❌ Login failed:${NC}"
  echo "$RESPONSE" | jq '.errors'
  exit 1
fi

# Extract data
USER_DATA=$(echo "$RESPONSE" | jq '.data.login.user')
ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.data.login.accessToken')
REFRESH_TOKEN=$(echo "$RESPONSE" | jq -r '.data.login.refreshToken')

echo -e "${GREEN}✅ Login successful!${NC}"
echo ""
echo "═══════════════════════════════════════════════════════════"
echo -e "${BLUE}USER DETAILS${NC}"
echo "═══════════════════════════════════════════════════════════"
echo "$USER_DATA" | jq '.'
echo ""
echo "═══════════════════════════════════════════════════════════"
echo -e "${BLUE}ACCESS TOKEN${NC}"
echo "═══════════════════════════════════════════════════════════"
echo "$ACCESS_TOKEN"
echo ""
echo "═══════════════════════════════════════════════════════════"
echo -e "${BLUE}REFRESH TOKEN${NC}"
echo "═══════════════════════════════════════════════════════════"
echo "$REFRESH_TOKEN"
echo ""

# Save tokens to file for later use
cat > /tmp/auth_tokens.json <<EOF
{
  "accessToken": "$ACCESS_TOKEN",
  "refreshToken": "$REFRESH_TOKEN",
  "user": $USER_DATA
}
EOF

echo -e "${GREEN}💾 Tokens saved to: /tmp/auth_tokens.json${NC}"
echo ""

# Test the token with DummyByID query
echo "═══════════════════════════════════════════════════════════"
echo -e "${BLUE}🧪 Testing DummyByID Query${NC}"
echo "═══════════════════════════════════════════════════════════"

DUMMY_QUERY='query DummyByID($id: Int!) {
  DummyByID(id: $id) {
    id
    name
  }
}'

DUMMY_RESPONSE=$(curl -s -X POST "$GATEWAY_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"query\": $(echo "$DUMMY_QUERY" | jq -Rs .),
    \"variables\": {
      \"id\": 1
    }
  }")

# Check for errors
if echo "$DUMMY_RESPONSE" | jq -e '.errors' > /dev/null 2>&1; then
  echo -e "${YELLOW}❌ DummyByID query failed:${NC}"
  echo "$DUMMY_RESPONSE" | jq '.errors'
else
  echo -e "${GREEN}✅ DummyByID query successful!${NC}"
  echo ""
  echo "$DUMMY_RESPONSE" | jq '.data.DummyByID'
fi

echo ""
echo -e "${YELLOW}📝 To use the access token in requests:${NC}"
echo -e "   ${BLUE}export TOKEN=\$(cat /tmp/auth_tokens.json | jq -r '.accessToken')${NC}"
echo -e "   ${BLUE}curl -H \"Authorization: Bearer \$TOKEN\" http://localhost:8080/graphql${NC}"

