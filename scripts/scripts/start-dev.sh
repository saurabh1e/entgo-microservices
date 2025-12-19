#!/usr/bin/env bash
set -euo pipefail

# Very small helper script to bring up the development compose stacks in order:
# 1) docker-compose.shared.yml (at repo root)
# 2) auth/docker-compose.dev.yml
# 3) All other docker-compose.dev.yml files found in the repo (excluding auth and gateway)
# 4) gateway/docker-compose.dev.yml
#
# Improvements:
# - Timestamped, colored headers for each step (can be disabled with --no-color)
# - Option --tail-logs N to show the last N lines of logs after a detached `up` (default 20)
# - Shows `docker compose ... ps` after each `up -d` to summarise container states
# - Option --follow-seconds N to follow combined logs for N seconds after a detached `up`
# - Auto-detects project root and navigates there

# Find the project root (directory containing docker-compose.shared.yml or go.work)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT=""

# Try to find project root by looking for marker files
if [[ -f "$SCRIPT_DIR/../docker-compose.shared.yml" ]] || [[ -f "$SCRIPT_DIR/../go.work" ]]; then
    PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
elif [[ -f "$SCRIPT_DIR/docker-compose.shared.yml" ]] || [[ -f "$SCRIPT_DIR/go.work" ]]; then
    PROJECT_ROOT="$SCRIPT_DIR"
else
    echo "Error: Could not find project root. Please run this script from the project root or scripts directory." >&2
    exit 1
fi

# Change to project root
cd "$PROJECT_ROOT" || {
    echo "Error: Failed to change to project root: $PROJECT_ROOT" >&2
    exit 1
}

echo "Working directory: $(pwd)"

# Function to check and create Docker network if it doesn't exist
ensure_network() {
  local network_name="$1"
  if ! docker network inspect "$network_name" >/dev/null 2>&1; then
    echo "Network '$network_name' not found. Creating it..."
    docker network create "$network_name" --driver bridge
    echo "✅ Network '$network_name' created successfully."
  else
    echo "✅ Network '$network_name' already exists."
  fi
}

# Function to check and create Docker volume if it doesn't exist
ensure_volume() {
  local volume_name="$1"
  if ! docker volume inspect "$volume_name" >/dev/null 2>&1; then
    echo "Volume '$volume_name' not found. Creating it..."
    docker volume create "$volume_name"
    echo "✅ Volume '$volume_name' created successfully."
  else
    echo "✅ Volume '$volume_name' already exists."
  fi
}

usage() {
  cat <<EOF
Usage: $0 [--no-detach] [--build] [--tail-logs N] [--follow-seconds N] [--no-color] [-h|--help]

Options:
  --no-detach       Run compose without -d (foreground)
  --build           Add --build to the compose up command
  --tail-logs N     Show the last N lines of logs after each detached compose up (default 20)
  --follow-seconds N Follow combined logs for N seconds after each detached compose up (default 0 = disabled)
  --no-color        Disable ANSI colors in output
  -h, --help        Show this help

Example: $0 --build --tail-logs 50 --follow-seconds 10
EOF
  exit 0
}

DETACH=true
BUILD=false
TAIL=20
FOLLOW_SECONDS=0
NO_COLOR=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --no-detach)
      DETACH=false; shift;;
    --build)
      BUILD=true; shift;;
    --tail-logs)
      shift; TAIL="$1"; shift;;
    --follow-seconds)
      shift; FOLLOW_SECONDS="$1"; shift;;
    --no-color)
      NO_COLOR=true; shift;;
    -h|--help)
      usage;;
    *)
      echo "Unknown option: $1"; usage;;
  esac
done

# Choose compose command: prefer "docker compose" (v2) but fall back to docker-compose (v1)
COMPOSE_CMD=""
if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
  COMPOSE_CMD=(docker compose)
elif command -v docker-compose >/dev/null 2>&1; then
  COMPOSE_CMD=(docker-compose)
else
  echo "Error: neither 'docker compose' nor 'docker-compose' found in PATH." >&2
  exit 2
fi

# Colors
if ! $NO_COLOR; then
  RESET="\033[0m"
  BOLD="\033[1m"
  CYAN="\033[36m"
  GREEN="\033[32m"
  YELLOW="\033[33m"
else
  RESET=""
  BOLD=""
  CYAN=""
  GREEN=""
  YELLOW=""
fi

timestamp() { date '+%Y-%m-%d %H:%M:%S' ; }

print_header() {
  local dir="$1" file="$2"
  echo -e "${CYAN}\n${BOLD}--- [$(timestamp)] Starting compose in: ${dir} (${file}) ---${RESET}"
}

print_footer() {
  local dur="$1"
  echo -e "${GREEN}${BOLD}--- Completed in ${dur}s ---${RESET}\n"
}

up_file() {
  local dir="$1" file="$2"
  print_header "$dir" "$file"

  pushd "$dir" >/dev/null || { echo "failed to cd to $dir" >&2; return 1; }

  # Build command args
  args=("${COMPOSE_CMD[@]}" -f "$file" up)
  if $BUILD; then args+=(--build); fi
  if $DETACH; then args+=(-d); fi

  echo "Running: ${args[*]}"

  start_ts=$(date +%s)
  # Run compose up. If running in foreground (no-detach) this will block until stopped.
  "${args[@]}"
  end_ts=$(date +%s)
  dur=$((end_ts - start_ts))

  # If we started in detached mode, provide a summary with ps and optional logs/follow
  if $DETACH; then
    echo -e "\n${YELLOW}Summary (compose ps):${RESET}"
    # show ps output; ignore errors so a failing ps doesn't kill the script
    "${COMPOSE_CMD[@]}" -f "$file" ps || true

    # Show the last N lines of logs if requested
    if [[ "$TAIL" =~ ^[0-9]+$ ]] && [[ "$TAIL" -gt 0 ]]; then
      echo -e "\n${YELLOW}Showing last ${TAIL} lines of logs for compose: ${dir}/${file}${RESET}"
      # Attempt to show logs; ignore errors
      "${COMPOSE_CMD[@]}" -f "$file" logs --no-color --tail="$TAIL" || true
    fi

    # Optionally follow combined logs for a short period to make it easier to spot errors
    if [[ "$FOLLOW_SECONDS" =~ ^[0-9]+$ ]] && [[ "$FOLLOW_SECONDS" -gt 0 ]]; then
      echo -e "\n${YELLOW}Following combined logs for ${FOLLOW_SECONDS}s (ctrl-C to interrupt)...${RESET}"
      # Start logs in background, then sleep for the desired duration, then kill the follower.
      "${COMPOSE_CMD[@]}" -f "$file" logs --no-color -f &
      LOG_PID=$!
      # Sleep but trap SIGINT so the user can cancel
      sleep "$FOLLOW_SECONDS" || true
      # Kill the background logs follower if still running
      if kill -0 "$LOG_PID" 2>/dev/null; then
        kill "$LOG_PID" 2>/dev/null || true
        wait "$LOG_PID" 2>/dev/null || true
      fi
    fi
  else
    echo -e "\n${YELLOW}Ran in foreground (no-detach); skipping ps/log tailing for this compose.${RESET}"
  fi

  print_footer "$dur"
  popd >/dev/null
}

# Ensure required networks and volumes exist
echo -e "\n${CYAN}${BOLD}--- Checking Docker networks and volumes ---${RESET}"

# Create the shared network that all services will use
ensure_network "entgo_network"

# Optionally create shared volumes (add more as needed)
# ensure_volume "redis_data"

echo -e "${GREEN}${BOLD}--- Network and volume checks complete ---${RESET}\n"

# 1) Start shared compose at repo root (if it exists)
if [[ -f "docker-compose.shared.yml" ]]; then
  up_file "." "docker-compose.shared.yml"
else
  echo "Warning: docker-compose.shared.yml not found in repo root; skipping." >&2
fi

# 2) Start auth/dev
if [[ -f "auth/docker-compose.dev.yml" ]]; then
  up_file "auth" "docker-compose.dev.yml"
else
  echo "Warning: auth/docker-compose.dev.yml not found; skipping auth." >&2
fi

# 3) Start other microservices which have docker-compose.dev.yml, excluding auth and gateway
echo -e "\nSearching for other docker-compose.dev.yml files..."
# Use find to discover files. Exclude auth and gateway paths.
while IFS= read -r -d $'\0' file; do
  dir=$(dirname "$file")
  # Skip repo root files and those in auth/gateway
  if [[ "$dir" == "auth" || "$dir" == "gateway" ]]; then
    continue
  fi
  up_file "$dir" "docker-compose.dev.yml"
done < <(find . -type f -name 'docker-compose.dev.yml' -not -path './auth/*' -not -path './gateway/*' -print0)

# 4) Start gateway/dev
if [[ -f "gateway/docker-compose.dev.yml" ]]; then
  up_file "gateway" "docker-compose.dev.yml"
else
  echo "Warning: gateway/docker-compose.dev.yml not found; skipping gateway." >&2
fi

echo -e "\n${BOLD}All requested compose files have been started.${RESET}\n"
