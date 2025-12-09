#!/bin/bash

# Setup gRPC files (copy from auth or create from template)
setup_grpc_files() {
    local service_dir=$1
    local auth_dir=$2
    local service_name=$3
    local base_module=$4

    # Try to copy from auth first
    if [ -f "$auth_dir/grpc/server.go" ] && [ -f "$auth_dir/grpc/interceptors.go" ]; then
        cp "$auth_dir/grpc/server.go" "$service_dir/grpc/"
        cp "$auth_dir/grpc/interceptors.go" "$service_dir/grpc/"

        # Update imports in server.go
        sed -i.bak "s|github.com/saurabh/entgo-microservices/auth|${base_module}/${service_name}|g" "$service_dir/grpc/server.go"
        sed -i.bak "s|\"auth/|\"${base_module}/${service_name}/|g" "$service_dir/grpc/server.go"

        # Remove auth-specific proto imports
        sed -i.bak '/userv1 "github.com\/saurabh\/entgo-microservices\/pkg\/proto\/user\/v1"/d' "$service_dir/grpc/server.go"
        sed -i.bak '/rolev1 "github.com\/saurabh\/entgo-microservices\/pkg\/proto\/role\/v1"/d' "$service_dir/grpc/server.go"
        sed -i.bak '/permissionv1 "github.com\/saurabh\/entgo-microservices\/pkg\/proto\/permission\/v1"/d' "$service_dir/grpc/server.go"
        sed -i.bak '/rolepermissionv1 "github.com\/saurabh\/entgo-microservices\/pkg\/proto\/rolepermission\/v1"/d' "$service_dir/grpc/server.go"

        # Remove auth-specific service registrations
        sed -i.bak '/RegisterUserServiceServer/d' "$service_dir/grpc/server.go"
        sed -i.bak '/RegisterRoleServiceServer/d' "$service_dir/grpc/server.go"
        sed -i.bak '/RegisterPermissionServiceServer/d' "$service_dir/grpc/server.go"
        sed -i.bak '/RegisterRolePermissionServiceServer/d' "$service_dir/grpc/server.go"
        sed -i.bak '/NewUserService/d' "$service_dir/grpc/server.go"
        sed -i.bak '/NewRoleService/d' "$service_dir/grpc/server.go"
        sed -i.bak '/NewPermissionService/d' "$service_dir/grpc/server.go"
        sed -i.bak '/NewRolePermissionService/d' "$service_dir/grpc/server.go"

        rm -f "$service_dir/grpc/server.go.bak"

        # Update interceptors.go
        sed -i.bak "s|github.com/saurabh/entgo-microservices/auth|${base_module}/${service_name}|g" "$service_dir/grpc/interceptors.go"
        sed -i.bak "s|\"auth/|\"${base_module}/${service_name}/|g" "$service_dir/grpc/interceptors.go"
        rm -f "$service_dir/grpc/interceptors.go.bak"

        log_success "Copied gRPC files from auth service"
        return 0
    fi

    log_warning "Could not copy from auth, creating minimal gRPC files"
    return 1
}

