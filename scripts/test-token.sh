#!/bin/bash
# Test authentication by fetching current user details using saved token

set -e

# Configuration
GATEWAY_URL="http://localhost:8080/graphql"
TOKEN_FILE="/tmp/auth_tokens.json"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Check if token file exists
if [ ! -f "$TOKEN_FILE" ]; then
  echo -e "${RED}❌ Token file not found: $TOKEN_FILE${NC}"
  echo -e "${YELLOW}   Run ./scripts/login.sh first${NC}"
  exit 1
fi

# Get access token
ACCESS_TOKEN=$(cat "$TOKEN_FILE" | jq -r '.accessToken')

if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" = "null" ]; then
  echo -e "${RED}❌ Invalid token in file${NC}"
  exit 1
fi

echo -e "${BLUE}🔍 Fetching user details with token...${NC}"
echo ""

# GraphQL query for "me"
QUERY='query Me {
  me {
    id
    username
    email
    name
    userType
    isActive
    emailVerified
    createdAt
    updatedAt
  }
}'

# Make the authenticated request
RESPONSE=$(curl -s -X POST "$GATEWAY_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d "{
    \"query\": $(echo "$QUERY" | jq -Rs .)
  }")

# Check for errors
if echo "$RESPONSE" | jq -e '.errors' > /dev/null 2>&1; then
  echo -e "${RED}❌ Request failed:${NC}"
  echo "$RESPONSE" | jq '.errors'
  exit 1
fi

# Extract user data
USER_DATA=$(echo "$RESPONSE" | jq '.data.me')

if [ "$USER_DATA" = "null" ]; then
  echo -e "${RED}❌ No user data returned. Token may be invalid or expired.${NC}"
  exit 1
fi

echo -e "${GREEN}✅ Authentication successful!${NC}"
echo ""
echo "═══════════════════════════════════════════════════════════"
echo -e "${BLUE}CURRENT USER${NC}"
echo "═══════════════════════════════════════════════════════════"
echo "$USER_DATA" | jq '.'
echo ""
echo -e "${GREEN}🎉 Token is valid and working!${NC}"

