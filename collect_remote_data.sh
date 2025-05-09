#!/usr/bin/env bash
set -e

# Helper functions
log_info() {
    echo "[INFO] $1"
}

log_error() {
    echo "[ERROR] $1" >&2
}

# Check required commands
for cmd in terraform jq rsync ssh; do
    if ! command -v "$cmd" &>/dev/null; then
        log_error "$cmd is required but not installed"
        exit 1
    fi
done

# Get SSH key path
SSH_KEY_PATH="./terraform/certs/ssh_key"

# Ensure the SSH key is not too open
if [[ ! -f "$SSH_KEY_PATH" ]]; then
    log_error "SSH key not found at $SSH_KEY_PATH"
    exit 1
fi
if [[ $(stat -c "%a" "$SSH_KEY_PATH") -gt 600 ]]; then
    log_error "SSH key permissions are too open. Please set to 600."
    exit 1
fi

# Create timestamp for this collection
DEPLOY_TIME_FILE="./terraform/deploy_time.txt"
if [[ ! -f "$DEPLOY_TIME_FILE" ]]; then
    log_error "Deployment time file not found: $DEPLOY_TIME_FILE"
    exit 1
fi
DEPLOY_TIME=$(cat "$DEPLOY_TIME_FILE")
#TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
STAT_DIR="./stat/$DEPLOY_TIME/$1"

# Create directories
mkdir -p "$STAT_DIR"

# Get service endpoints
ENDPOINTS=$(terraform -chdir=$(dirname "$0")/terraform output -json service_endpoints) || {
    log_error "Failed to get service endpoints from terraform"
    exit 1
}

# Validate JSON output
echo "$ENDPOINTS" | jq empty 2>/dev/null || {
    log_error "Invalid JSON output from terraform"
    log_error "Raw output: $ENDPOINTS"
    exit 1
}

# Process each endpoint
# https://chatgpt.com/share/681d8185-2ce0-8000-be23-a8eff8217981
mapfile -t LINES < <(echo "$ENDPOINTS" | jq -r 'to_entries[] | "\(.key)|\(.value)"')
for LINE in "${LINES[@]}"; do
    IFS='|' read -r SERVICE ENDPOINT <<<"$LINE"

    if [[ -z "$SERVICE" || -z "$ENDPOINT" ]]; then
        log_error "Invalid service or endpoint found"
        continue
    fi

    log_info "Collecting data from $SERVICE at $ENDPOINT"

    # Create service directory
    mkdir -p "$STAT_DIR/$SERVICE"

    # Test SSH connection first
    if ! ssh -i "$SSH_KEY_PATH" -o StrictHostKeyChecking=no -o ConnectTimeout=10 "ec2-user@$ENDPOINT" "exit" 2>/dev/null; then
        log_error "Cannot connect to $SERVICE at $ENDPOINT"
        continue
    fi

    # Attempt to copy stats directory
    if ssh -i "$SSH_KEY_PATH" -o StrictHostKeyChecking=no "ec2-user@$ENDPOINT" "test -d /opt/flappygo/backend/statout"; then
        rsync -az --timeout=30 -e "ssh -i $SSH_KEY_PATH -o StrictHostKeyChecking=no" \
            "ec2-user@$ENDPOINT:/opt/flappygo/backend/statout/" \
            "$STAT_DIR/$SERVICE" || {
            log_error "Failed to collect stats from $SERVICE"
            continue
        }
        log_info "Successfully collected stats from $SERVICE"
    else
        log_error "Stats directory not found on $SERVICE"
    fi
done

log_info "Data collection completed. Stats saved in $STAT_DIR"
