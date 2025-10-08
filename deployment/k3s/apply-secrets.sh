#!/bin/bash

# Script to apply Kubernetes secrets using environment variables from .env file
# This script reads .env file and substitutes variables in secret.yaml

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🔐 Applying Kubernetes Secrets for Belimang Application${NC}"

# Get project root directory
CURRENT_DIR=$(pwd)
PROJECT_ROOT=$(dirname $(dirname "$CURRENT_DIR"))

# Check if root .env file exists
if [ ! -f "$PROJECT_ROOT/.env" ]; then
    echo -e "${RED}❌ Error: .env file not found in project root!${NC}"
    echo -e "${YELLOW}💡 Please copy .env.example.k3s to .env in project root and update the values${NC}"
    exit 1
fi

# Load environment variables from root .env file
echo -e "${YELLOW}📄 Loading environment variables from root .env${NC}"
export $(grep -v '^#' "$PROJECT_ROOT/.env" | xargs)

# Check if required variables are set
required_vars=("DB_PASSWORD" "JWT_SECRET_KEY" "DATABASE_URL")
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo -e "${RED}❌ Error: Required environment variable $var is not set in k8s/.env${NC}"
        exit 1
    fi
done

# Create temporary file with substituted values
temp_secret=$(mktemp)
envsubst < secret.yaml > "$temp_secret"

echo -e "${YELLOW}🚀 Applying secret to Kubernetes cluster${NC}"

# Apply the secret
if kubectl apply -f "$temp_secret"; then
    echo -e "${GREEN}✅ Secret 'belimang-secret' applied successfully!${NC}"
else
    echo -e "${RED}❌ Failed to apply secret${NC}"
    rm "$temp_secret"
    exit 1
fi

# Clean up temporary file
rm "$temp_secret"

echo -e "${GREEN}🎉 Secrets deployment completed!${NC}"