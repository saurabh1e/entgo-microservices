#!/usr/bin/env bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default old module path
OLD_MODULE_PATH="github.com/saurabh/entgo-microservices"

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to display usage
usage() {
    cat << EOF
Usage: $0 <new-module-path>

Replace the Go module path throughout the project.

Arguments:
    new-module-path    The new module path (e.g., github.com/yourname/yourproject)

Example:
    $0 github.com/mycompany/myproject

This script will:
    1. Update all go.mod files
    2. Update all .go source files
    3. Update all .proto files
    4. Regenerate .pb.go files with new paths
    5. Update go.work file if present

EOF
    exit 1
}

# Check if the new module path is provided
if [ $# -ne 1 ]; then
    print_error "Missing required argument: new-module-path"
    usage
fi

NEW_MODULE_PATH="$1"

# Validate new module path format
if [[ ! "$NEW_MODULE_PATH" =~ ^[a-zA-Z0-9._-]+(/[a-zA-Z0-9._-]+)+$ ]]; then
    print_error "Invalid module path format: $NEW_MODULE_PATH"
    print_info "Module path should be in format: domain.com/org/project"
    exit 1
fi

# Get the project root directory (parent of the scripts directory)
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

print_info "Project root: $PROJECT_ROOT"
print_info "Old module path: $OLD_MODULE_PATH"
print_info "New module path: $NEW_MODULE_PATH"
echo ""

# Confirmation prompt
read -p "$(echo -e ${YELLOW}Are you sure you want to proceed? This will modify multiple files. [y/N]:${NC} )" -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_warning "Operation cancelled by user"
    exit 0
fi

echo ""
print_info "Starting module path replacement..."
echo ""

# Counter for modified files
TOTAL_FILES=0

# 1. Update go.mod files
print_info "Step 1: Updating go.mod files..."
GO_MOD_FILES=$(find "$PROJECT_ROOT" -name "go.mod" -type f)
for file in $GO_MOD_FILES; do
    if grep -q "$OLD_MODULE_PATH" "$file"; then
        print_info "  Updating: $file"
        # Use temporary file for safe replacement
        if [[ "$OSTYPE" == "darwin"* ]]; then
            # macOS
            sed -i '' "s|$OLD_MODULE_PATH|$NEW_MODULE_PATH|g" "$file"
        else
            # Linux
            sed -i "s|$OLD_MODULE_PATH|$NEW_MODULE_PATH|g" "$file"
        fi
        ((TOTAL_FILES++))
    fi
done
print_success "Updated go.mod files"
echo ""

# 2. Update .go source files (excluding vendor and generated pb.go files initially)
print_info "Step 2: Updating .go source files..."
GO_FILES=$(find "$PROJECT_ROOT" -name "*.go" -type f \
    ! -path "*/vendor/*" \
    ! -path "*/node_modules/*" \
    ! -path "*/.git/*")

for file in $GO_FILES; do
    if grep -q "$OLD_MODULE_PATH" "$file"; then
        print_info "  Updating: $file"
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s|$OLD_MODULE_PATH|$NEW_MODULE_PATH|g" "$file"
        else
            sed -i "s|$OLD_MODULE_PATH|$NEW_MODULE_PATH|g" "$file"
        fi
        ((TOTAL_FILES++))
    fi
done
print_success "Updated .go source files"
echo ""

# 3. Update .proto files
print_info "Step 3: Updating .proto files..."
PROTO_FILES=$(find "$PROJECT_ROOT" -name "*.proto" -type f)
for file in $PROTO_FILES; do
    if grep -q "$OLD_MODULE_PATH" "$file"; then
        print_info "  Updating: $file"
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s|$OLD_MODULE_PATH|$NEW_MODULE_PATH|g" "$file"
        else
            sed -i "s|$OLD_MODULE_PATH|$NEW_MODULE_PATH|g" "$file"
        fi
        ((TOTAL_FILES++))
    fi
done
print_success "Updated .proto files"
echo ""

# 4. Update go.work file if it exists
print_info "Step 4: Checking for go.work file..."
if [ -f "$PROJECT_ROOT/go.work" ]; then
    if grep -q "$OLD_MODULE_PATH" "$PROJECT_ROOT/go.work"; then
        print_info "  Updating: $PROJECT_ROOT/go.work"
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s|$OLD_MODULE_PATH|$NEW_MODULE_PATH|g" "$PROJECT_ROOT/go.work"
        else
            sed -i "s|$OLD_MODULE_PATH|$NEW_MODULE_PATH|g" "$PROJECT_ROOT/go.work"
        fi
        ((TOTAL_FILES++))
    fi
    print_success "Updated go.work file"
else
    print_info "  No go.work file found (skipping)"
fi
echo ""

# 5. Update any other configuration files that might contain the module path
print_info "Step 5: Checking for other configuration files..."
CONFIG_PATTERNS=("*.yml" "*.yaml" "*.json" "Makefile" "Dockerfile" "*.mk" "*.sh")
for pattern in "${CONFIG_PATTERNS[@]}"; do
    while IFS= read -r -d '' file; do
        if grep -q "$OLD_MODULE_PATH" "$file" 2>/dev/null; then
            print_info "  Updating: $file"
            if [[ "$OSTYPE" == "darwin"* ]]; then
                sed -i '' "s|$OLD_MODULE_PATH|$NEW_MODULE_PATH|g" "$file"
            else
                sed -i "s|$OLD_MODULE_PATH|$NEW_MODULE_PATH|g" "$file"
            fi
            ((TOTAL_FILES++))
        fi
    done < <(find "$PROJECT_ROOT" -name "$pattern" -type f \
        ! -path "*/vendor/*" \
        ! -path "*/node_modules/*" \
        ! -path "*/.git/*" \
        -print0 2>/dev/null)
done
print_success "Updated configuration files"
echo ""

# Summary
echo ""
print_success "================================"
print_success "Module path replacement complete!"
print_success "================================"
echo ""
print_info "Total files modified: $TOTAL_FILES"
echo ""
print_warning "Next steps:"
echo "  1. Regenerate protocol buffer files if needed:"
echo "     ${BLUE}cd proto && make${NC}"
echo ""
echo "  2. Tidy up Go modules:"
echo "     ${BLUE}go work sync${NC}"
echo "     ${BLUE}cd auth && go mod tidy${NC}"
echo "     ${BLUE}cd gateway && go mod tidy${NC}"
echo "     ${BLUE}cd pkg && go mod tidy${NC}"
echo "     ${BLUE}cd test-service && go mod tidy${NC}"
echo ""
echo "  3. Test your services to ensure everything works:"
echo "     ${BLUE}make test${NC}"
echo ""
print_info "Review the changes with: ${BLUE}git diff${NC}"

