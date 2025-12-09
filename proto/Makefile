.PHONY: all clean proto

# Proto generation
PROTO_DIR := .
PKG_PROTO_DIR := ../pkg/proto
PROTOC := protoc
PROTOC_GEN_GO := $(shell go env GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(shell go env GOPATH)/bin/protoc-gen-go-grpc

# Find all proto files dynamically
PROTO_FILES := $(shell find $(PROTO_DIR) -name "*.proto" -type f)

# Install protoc plugins if not present
install-tools:
	@which protoc > /dev/null || (echo "protoc not found. Install from https://grpc.io/docs/protoc-installation/" && exit 1)
	@test -f $(PROTOC_GEN_GO) || go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@test -f $(PROTOC_GEN_GO_GRPC) || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate all proto files
all: install-tools proto

# Clean generated files
clean:
	find $(PKG_PROTO_DIR) -name "*.pb.go" -delete

# Generate proto files dynamically
proto:
	@echo "Generating proto files..."
	@for proto_file in $(PROTO_FILES); do \
		echo "Compiling $$proto_file..."; \
		$(PROTOC) --proto_path=$(PROTO_DIR) \
			--go_out=$(PKG_PROTO_DIR) \
			--go_opt=module=github.com/saurabh/entgo-microservices/pkg \
			--go-grpc_out=$(PKG_PROTO_DIR) \
			--go-grpc_opt=module=github.com/saurabh/entgo-microservices/pkg \
			$$proto_file || exit 1; \
	done
	@echo "✅ Proto generation complete!"



