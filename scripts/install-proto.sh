#!/usr/bin/env bash
set -euo pipefail

# install-proto.sh
# Installs protoc and common Go protobuf plugins on macOS/Linux.
# On macOS this uses Homebrew for protoc. On Linux it will attempt apt-get (Debian/Ubuntu).
# It also installs the Go plugins protoc-gen-go and protoc-gen-go-grpc using `go install`.

usage() {
  cat <<EOF
Usage: $0 [--run]

This script will:
  - Check for protoc; if missing, attempt to install it (brew on macOS, apt on Debian/Ubuntu).
  - Install Go protoc plugins: protoc-gen-go and protoc-gen-go-grpc (using 'go install').

By default the script only performs checks and prints the commands it would run.
Pass --run to actually perform installations.

Example: $0 --run
EOF
}

DRY_RUN=true
if [[ ${1:-} == "--run" ]]; then
  DRY_RUN=false
fi

echo "Checking environment for protoc and Go protobuf plugins..."

run_cmd() {
  if $DRY_RUN; then
    echo "DRY RUN: $*"
  else
    echo "Running: $*"
    # execute the command string in the current shell
    bash -lc "$*"
  fi
}

# 1) Check protoc
if command -v protoc >/dev/null 2>&1; then
  echo "protoc found: $(protoc --version)"
else
  echo "protoc not found."
  if [[ "$DRY_RUN" == "true" ]]; then
    if [[ "$(uname)" == "Darwin" ]]; then
      echo "Would install via: brew install protobuf"
    else
      echo "On Debian/Ubuntu you'd run: sudo apt-get update && sudo apt-get install -y protobuf-compiler"
    fi
  else
    if [[ "$(uname)" == "Darwin" ]]; then
      if command -v brew >/dev/null 2>&1; then
        run_cmd "brew update && brew install protobuf"
      else
        echo "Homebrew not found. Please install Homebrew first: https://brew.sh/" >&2
        exit 1
      fi
    else
      # Attempt apt-get
      if command -v apt-get >/dev/null 2>&1; then
        run_cmd "sudo apt-get update && sudo apt-get install -y protobuf-compiler"
      else
        echo "Unsupported OS or package manager. Please install protoc manually." >&2
        exit 1
      fi
    fi
  fi
fi

# 2) Install Go plugins (protoc-gen-go and protoc-gen-go-grpc)
# These install into $(go env GOPATH)/bin or $GOBIN if set. Ensure that directory is on your PATH.

if ! command -v go >/dev/null 2>&1; then
  echo "Go not found in PATH. Please install Go (1.20+ recommended)." >&2
  exit 2
fi

PROTOC_GEN_GO_MODULE="google.golang.org/protobuf/cmd/protoc-gen-go@latest"
PROTOC_GEN_GO_GRPC_MODULE="google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"

echo "\nWill install Go plugins: protoc-gen-go and protoc-gen-go-grpc"
run_cmd "go install ${PROTOC_GEN_GO_MODULE}"
run_cmd "go install ${PROTOC_GEN_GO_GRPC_MODULE}"

# 3) Advice about PATH
GOBIN_DIR="$(go env GOPATH 2>/dev/null || echo \"$HOME/go\")/bin"
if [[ -n "$(go env GOBIN 2>/dev/null)" ]]; then
  GOBIN_DIR="$(go env GOBIN)"
fi

echo "\nPost-install notes:"
echo "  - Ensure \"$GOBIN_DIR\" is on your PATH so protoc can find protoc-gen-go and protoc-gen-go-grpc."
echo "    e.g. add to ~/.zshrc or ~/.profile: export PATH=\"$GOBIN_DIR:\$PATH\""

if $DRY_RUN; then
  echo "\nDry run mode: nothing was changed. Re-run with --run to perform the installations."
else
  echo "\nInstallation steps completed (if no errors were reported)."
  echo "Verify by running: protoc --version && protoc-gen-go --version || echo 'protoc-gen-go binary may not support --version'"
fi
