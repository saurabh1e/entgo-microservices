#!/bin/bash

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Logging functions
log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

# Validation functions
validate_service_name() {
    local service_name=$1

    if [ -z "$service_name" ]; then
        log_error "Service name cannot be empty"
        return 1
    fi

    # Check if service name contains only valid characters (lowercase, numbers, hyphens)
    if ! [[ "$service_name" =~ ^[a-z0-9-]+$ ]]; then
        log_error "Service name must contain only lowercase letters, numbers, and hyphens"
        return 1
    fi

    return 0
}

validate_directory_not_exists() {
    local dir=$1

    if [ -d "$dir" ]; then
        log_error "Directory '$dir' already exists"
        return 1
    fi

    return 0
}

# String manipulation utilities
to_env_var_name() {
    local name=$1
    echo "$name" | tr '[:lower:]' '[:upper:]' | tr '-' '_'
}

capitalize_first() {
    local str=$1

    local first="${str:0:1}"
    local rest="${str:1}"
    echo "$(echo "$first" | tr '[:lower:]' '[:upper:]')${rest}"
}

