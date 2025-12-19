#!/bin/bash

# Integrate new service with gateway
integrate_with_gateway() {
    local service_name=$1
    local service_port=$2
    local gateway_dir=$3
    local service_name_upper=$4

    if [ ! -d "$gateway_dir" ]; then
        log_warning "Gateway directory not found at $gateway_dir, skipping gateway integration"
        return 1
    fi

    # Add service to services.conf file
    if [ -f "$gateway_dir/services.conf" ]; then
        # Check if service is already registered
        if ! grep -q "^${service_name}|" "$gateway_dir/services.conf" 2>/dev/null; then
            echo "${service_name}|http://entgo_${service_name}_dev:${service_port}/graphql" >> "$gateway_dir/services.conf"
            log_success "Added $service_name to gateway/services.conf"
        else
            log_warning "$service_name already exists in gateway/services.conf"
        fi
    else
        # Create services.conf if it doesn't exist
        cat > "$gateway_dir/services.conf" << EOF
# Gateway Service Registry
# This file lists all microservices that should be registered with the gateway
# Format: service-name|service-url
# Lines starting with # are ignored

# Core services
auth|http://entgo_auth_dev:8081/graphql
${service_name}|http://entgo_${service_name}_dev:${service_port}/graphql
EOF
        log_success "Created gateway/services.conf with $service_name"
    fi

    log_success "Gateway integration completed - service registered in services.conf"
    return 0
}

