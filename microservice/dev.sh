#!/bin/bash

# Discovery App Development Script
# This script helps with building Ent, generating GraphQL, and running the server

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Cross-platform sed in-place editing
sed_i() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "$@"
    else
        # Linux
        sed -i "$@"
    fi
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  clean        - Clean all generated files and binaries"
    echo "  run          - Regenerate GraphQL and run using main.go"
    echo "  dev          - Regenerate GraphQL and run with hot reload using Air"
    echo "  build        - Clean, build everything, and create executable"
    echo "  help         - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 clean     - Clean everything"
    echo "  $0 run       - Regenerate and run with go run main.go"
    echo "  $0 dev       - Development mode with hot reload"
    echo "  $0 build     - Full build with executable"
}

# Function to ensure dependencies are installed once
ensure_dependencies() {
    log_info "📦 Ensuring all dependencies are available..."

    # Only install dependencies if they're not already cached
    if [ ! -f ".deps_installed" ]; then
        log_info "🔧 Installing required dependencies..."

        # Install ent
        go get -u entgo.io/ent/cmd/ent

        # Install gqlgen with specific version
        go get github.com/99designs/gqlgen@v0.17.76

        # Mark dependencies as installed
        touch .deps_installed
        log_info "✅ Dependencies installed and cached"
    else
        log_info "✅ Dependencies already installed (cached)"
    fi

    # Tidy up go modules
    if go mod tidy; then
        log_info "✅ Go modules tidied"
    else
        log_warning "⚠️  Go mod tidy had issues, continuing..."
    fi
}

# Function to build everything using generate.go (which handles hooks, privacy, resolvers)
build_all() {
    log_info "🔧 Generating all code using generate.go..."

    # Ensure dependencies are available
    ensure_dependencies

    # Create internal directory if it doesn't exist
    mkdir -p internal

    # Run the comprehensive generate.go which handles everything
    if go generate generate.go; then
        log_success "✅ All code generated successfully"

        # Verify that internal/ent was created
        if [ ! -d "internal/ent" ]; then
            log_error "❌ Code generation didn't create internal/ent directory"
            log_info "💡 Check your generate.go file configuration"
            exit 1
        fi

        # Run go mod tidy after generation to resolve dependencies
        log_info "📦 Tidying Go modules..."
        if go mod tidy; then
            log_success "✅ Go modules tidied successfully"
        else
            log_warning "⚠️  Warning: go mod tidy had issues (may be okay in workspace mode)"
        fi
    else
        log_error "❌ Failed to generate code"
        exit 1
    fi
}

# Function to generate only GraphQL
generate_graphql_only() {
    log_info "🎯 Generating GraphQL code..."

    if go run -mod=mod github.com/99designs/gqlgen; then
        log_success "✅ GraphQL code generated successfully"
    else
        log_error "❌ Failed to generate GraphQL code"
        exit 1
    fi
}

# Function to build the application executable
build_app() {
    log_info "🏗️  Building application executable..."
    if go build -o bin/microservice .; then
        log_success "✅ Application built successfully"
    else
        log_error "❌ Failed to build application"
        exit 1
    fi
}

# Function to run the server with go run main.go
run_server() {
    log_info "🚀 Starting server with go run main.go..."
    go run main.go
}

# Function to clean schema methods and imports (Policy and Hooks)
clean_schema_methods() {
    log_info "🧹 Cleaning Policy(), Hooks() methods and imports from schemas..."

    SCHEMA_DIR="./ent/schema"
    shopt -s nullglob 2>/dev/null || true

    for schema_file in "$SCHEMA_DIR"/*.go; do
        [[ -f "$schema_file" ]] || continue

        base="$(basename "$schema_file")"
        # Skip mixin and enum files
        if [[ "$base" == *"mixin"* || "$base" == *"enum"* ]]; then
            continue
        fi

        # Remove Policy() method if present (using perl for better compatibility)
        if grep -Eq '^func \([A-Za-z_][A-Za-z0-9_]*\) Policy\(\) ent\.Policy' "$schema_file"; then
            log_info "  Removing Policy() method from $base"
            perl -i -ne 'print unless /^func \([A-Za-z_][A-Za-z0-9_]*\) Policy\(\) ent\.Policy/ .. /^}/' "$schema_file"
        fi

        # Remove Hooks() method if present (using perl for better compatibility)
        if grep -Eq '^func \([A-Za-z_][A-Za-z0-9_]*\) Hooks\(\) \[\]ent\.Hook' "$schema_file"; then
            log_info "  Removing Hooks() method from $base"
            perl -i -ne 'print unless /^func \([A-Za-z_][A-Za-z0-9_]*\) Hooks\(\) \[\]ent\.Hook/ .. /^}/' "$schema_file"
        fi

        # Clean up specific imports only (keep the rest intact)
        # This handles both short paths and full module paths
        log_info "  Cleaning imports in $base"
        sed_i \
            -e '/^[[:space:]]*hooks ".*\/ent\/schema_hooks"$/d' \
            -e '/^[[:space:]]*hook ".*\/ent\/schema_hooks"$/d' \
            -e '/^[[:space:]]*".*\/ent\/hooks"$/d' \
            -e '/^[[:space:]]*privacy ".*\/ent\/schema_privacy"$/d' \
            -e '/^[[:space:]]*".*\/ent\/privacy"$/d' \
            -e '/^[[:space:]]*".*\/ent\/schema_privacy"$/d' \
            "$schema_file"

        # Also remove any imports that match the pattern with aliasing
        # e.g., hook "github.com/user/project/microservice/ent/schema_hooks"
        sed_i \
            -e '/^[[:space:]]*[a-zA-Z_][a-zA-Z0-9_]* ".*\/schema_hooks"$/d' \
            -e '/^[[:space:]]*[a-zA-Z_][a-zA-Z0-9_]* ".*\/schema_privacy"$/d' \
            "$schema_file"

        # Run goimports to clean up any remaining unused imports
        if command -v goimports &> /dev/null; then
            goimports -w "$schema_file" 2>/dev/null || true
        elif command -v gofmt &> /dev/null; then
            # Fallback to gofmt if goimports is not available
            gofmt -w "$schema_file" 2>/dev/null || true
        fi
    done

    log_success "✅ Finished cleaning Policy(), Hooks() methods and imports from schemas"
}

# Function to clean ALL generated files
clean_all() {
    log_info "🧹 Cleaning all generated files and binaries..."

    # Clean schema methods and imports first
    clean_schema_methods

    # Remove Ent generated files
    if [ -d "internal/ent" ]; then
        rm -rf internal/ent
        log_info "✅ Removed internal/ent directory"
    fi

    # Remove GraphQL generated files
    if [ -f "graph/generated.go" ]; then
        rm -f graph/generated.go
        log_info "✅ Removed graph/generated.go"
    fi

    if [ -d "graph/model" ]; then
        rm -rf graph/model
        log_info "✅ Removed graph/model directory"
    fi

    # NOTE: We DO NOT delete ent/schema_hooks/ and ent/schema_privacy/ directories
    # These contain your custom implementations that should be preserved!

    # Remove binary and bin directory
    if [ -d "bin" ]; then
        rm -rf bin
        log_info "✅ Removed bin directory"
    fi

    # Remove temporary files
    if [ -d "tmp" ]; then
        rm -rf tmp
        log_info "✅ Removed tmp directory"
    fi

    # Remove log files
    if [ -f "build-errors.log" ]; then
        rm -f build-errors.log
        log_info "✅ Removed build-errors.log"
    fi

    # Remove dependency cache
    if [ -f ".deps_installed" ]; then
        rm -f .deps_installed
        log_info "✅ Removed dependency cache"
    fi

    log_success "✅ All cleanup completed - custom hooks and privacy files preserved"
}

# Function to regenerate and run
regenerate_and_run() {
    log_info "🔄 Regenerating GraphQL and running server..."

    # Generate only GraphQL
    generate_graphql_only

    # Run server
    run_server
}

# Function for full build process
full_build() {
    log_info "🔄 Starting full build process..."

    # Create bin directory if it doesn't exist
    mkdir -p bin

    # Step 1: Clean everything first
    log_info "🧹 Phase 1: Cleaning all files..."
    clean_all

    # Step 2: Generate all code (Ent, GraphQL, hooks, privacy, resolvers)
    log_info "🔧 Phase 2: Generating all code..."
    build_all

    # Step 3: Build application executable
    log_info "🏗️  Phase 3: Building application executable..."
    build_app

    log_success "🎉 Full build completed successfully!"
}

# Function for development mode with file watching
dev_mode() {
    log_info "🔧 Starting development mode with hot reload..."

    # Get the Go bin path
    GOBIN=$(go env GOBIN)
    if [ -z "$GOBIN" ]; then
        GOPATH=$(go env GOPATH)
        GOBIN="$GOPATH/bin"
    fi

    # Check if air is installed (use full path)
    if [ ! -f "$GOBIN/air" ]; then
        log_warning "Air not found. Installing air for hot reload..."
        go install github.com/air-verse/air@latest

        # Verify installation
        if [ ! -f "$GOBIN/air" ]; then
            log_error "❌ Failed to install Air. Please check your Go installation and PATH."
            exit 1
        fi
    fi

    # Create .air.toml config if it doesn't exist
    if [ ! -f ".air.toml" ]; then
        log_info "Creating .air.toml configuration..."
        cat > .air.toml << 'EOF'
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go run main.go"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "internal/ent", "graph/generated.go"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html", "graphqls"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  rerun = true
  rerun_delay = 500
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
EOF
    fi

    # Run go mod tidy before starting
    log_info "📦 Running go mod tidy..."
    if go mod tidy; then
        log_success "✅ Go modules tidied successfully"
    else
        log_warning "⚠️  Warning: go mod tidy had issues (may be okay in workspace mode)"
    fi

    # Generate all code before starting air
    log_info "🔄 Generating all code with generate.go..."
    if go generate generate.go; then
        log_success "✅ Code generation completed successfully"
    else
        log_error "❌ Failed to generate code"
        exit 1
    fi

    # Start air for hot reload using full path
    log_info "🔥 Starting hot reload with Air..."
    "$GOBIN/air"
}

# Function to check prerequisites
check_prerequisites() {
    log_info "🔍 Checking prerequisites..."

    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi

    # Check if we're in the right directory
    if [ ! -f "go.mod" ] || [ ! -f "main.go" ]; then
        log_error "This doesn't appear to be the root of themicroservice project"
        exit 1
    fi

    log_success "✅ Prerequisites check passed"
}

# Main script logic
main() {
    # Check prerequisites first
    check_prerequisites

    case "${1:-help}" in
        "clean")
            clean_all
            ;;
        "run")
            regenerate_and_run
            ;;
        "dev")
            dev_mode
            ;;
        "build")
            full_build
            ;;
        "help"|"-h"|"--help")
            show_usage
            ;;
        *)
            log_error "Unknown command: $1"
            echo ""
            show_usage
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
