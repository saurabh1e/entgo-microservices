#!/bin/bash

# Copy and update files from auth service
copy_from_auth() {
    local source_dir=$1
    local dest_dir=$2
    local service_name=$3
    local base_module=$4

    # Copy cmd folder (excluding seed)
    if [ -d "$source_dir/cmd" ]; then
        cp -r "$source_dir/cmd" "$dest_dir/"
        rm -rf "$dest_dir/cmd/seed" 2>/dev/null

        # Update imports in cmd files (both relative and full module paths)
        find "$dest_dir/cmd" -name "*.go" -exec sed -i.bak "s|\"auth/|\"${base_module}/${service_name}/|g" {} \;
        find "$dest_dir/cmd" -name "*.go" -exec sed -i.bak "s|github.com/saurabh/entgo-microservices/auth|${base_module}/${service_name}|g" {} \;
        find "$dest_dir/cmd" -name "*.bak" -delete
    fi

    # Copy utils folder
    if [ -d "$source_dir/utils" ]; then
        cp -r "$source_dir/utils" "$dest_dir/"

        # Update imports in utils files (both relative and full module paths)
        find "$dest_dir/utils" -name "*.go" -exec sed -i.bak "s|\"auth/|\"${base_module}/${service_name}/|g" {} \;
        find "$dest_dir/utils" -name "*.go" -exec sed -i.bak "s|github.com/saurabh/entgo-microservices/auth|${base_module}/${service_name}|g" {} \;
        find "$dest_dir/utils" -name "*.bak" -delete
    fi

    # Copy config folder
    if [ -d "$source_dir/config" ]; then
        cp -r "$source_dir/config" "$dest_dir/"

        # Update imports in config files (both relative and full module paths)
        find "$dest_dir/config" -name "*.go" -exec sed -i.bak "s|\"auth/|\"${base_module}/${service_name}/|g" {} \;
        find "$dest_dir/config" -name "*.go" -exec sed -i.bak "s|github.com/saurabh/entgo-microservices/auth|${base_module}/${service_name}|g" {} \;
        find "$dest_dir/config" -name "*.bak" -delete
    fi
}

# Copy configuration files
copy_config_files() {
    local source_dir=$1
    local dest_dir=$2
    local service_name=$3

    # Copy .air.toml
    [ -f "$source_dir/.air.toml" ] && cp "$source_dir/.air.toml" "$dest_dir/"

    # Copy .dockerignore
    [ -f "$source_dir/.dockerignore" ] && cp "$source_dir/.dockerignore" "$dest_dir/"

    # Copy and update .gitignore
    if [ -f "$source_dir/.gitignore" ]; then
        cp "$source_dir/.gitignore" "$dest_dir/"
        # Ensure logs/ and tmp/ are ignored
        grep -q "^logs/$" "$dest_dir/.gitignore" || echo "logs/" >> "$dest_dir/.gitignore"
        grep -q "^tmp/$" "$dest_dir/.gitignore" || echo "tmp/" >> "$dest_dir/.gitignore"
    fi
}

# Copy and update main.go
copy_main_go() {
    local source_file=$1
    local dest_file=$2
    local service_name=$3
    local base_module=$4

    if [ -f "$source_file" ]; then
        cp "$source_file" "$dest_file"

        # Update imports to use full module path
        sed -i.bak "s|github.com/saurabh/entgo-microservices/auth|${base_module}/${service_name}|g" "$dest_file"
        # Also replace relative imports
        sed -i.bak "s|\"auth/|\"${base_module}/${service_name}/|g" "$dest_file"
        # Replace import _ "auth/internal/ent/runtime"
        sed -i.bak "s|import _ \"auth/internal/ent/runtime\"|import _ \"${base_module}/${service_name}/internal/ent/runtime\"|g" "$dest_file"

        rm -f "${dest_file}.bak"
    fi
}

# Copy and update graph files
copy_graph_files() {
    local source_dir=$1
    local dest_dir=$2
    local service_name=$3
    local base_module=$4

    # Copy schema.graphqls
    if [ -f "$source_dir/graph/schema.graphqls" ]; then
        cp "$source_dir/graph/schema.graphqls" "$dest_dir/graph/schema.graphqls"
    fi

    # Copy and update resolver.go
    if [ -f "$source_dir/graph/resolver.go" ]; then
        cp "$source_dir/graph/resolver.go" "$dest_dir/graph/"
        sed -i.bak "s|\"auth/|\"${base_module}/${service_name}/|g" "$dest_dir/graph/resolver.go"
        sed -i.bak "s|github.com/saurabh/entgo-microservices/auth|${base_module}/${service_name}|g" "$dest_dir/graph/resolver.go"
        rm -f "$dest_dir/graph/resolver.go.bak"
    fi
}

