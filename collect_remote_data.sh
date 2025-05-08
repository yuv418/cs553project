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
    if ! command -v "$cmd" &> /dev/null; then
        log_error "$cmd is required but not installed"
        exit 1
    fi
done

# Get SSH key path
SSH_KEY_PATH="./certs/ssh_key"

# Create timestamp for this collection
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
STAT_DIR="stat/$TIMESTAMP"

# Create directories
mkdir -p "$STAT_DIR"

# Get service endpoints
cd "$(dirname "$0")/terraform"
ENDPOINTS=$(terraform output -json service_endpoints) || {
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
echo "$ENDPOINTS" | jq -r 'to_entries[] | "\(.key)|\(.value)"' | while IFS='|' read -r SERVICE ENDPOINT; do
    if [[ -z "$SERVICE" || -z "$ENDPOINT" ]]; then
        log_error "Invalid service or endpoint found"
        continue
    fi
    
    log_info "Collecting data from $SERVICE at $ENDPOINT"
    
    # Create service directory
    mkdir -p "../$STAT_DIR/$SERVICE"
    
    # Test SSH connection first
    if ! ssh -i "$SSH_KEY_PATH" -o StrictHostKeyChecking=no -o ConnectTimeout=10 "ec2-user@$ENDPOINT" "exit" 2>/dev/null; then
        log_error "Cannot connect to $SERVICE at $ENDPOINT"
        continue
    fi
    
    # Attempt to copy stats directory
    if ssh -i "$SSH_KEY_PATH" -o StrictHostKeyChecking=no "ec2-user@$ENDPOINT" "test -d /flappygo/backend/statout"; then
        rsync -az --timeout=30 -e "ssh -i $SSH_KEY_PATH -o StrictHostKeyChecking=no" \
            "ec2-user@$ENDPOINT:/flappygo/backend/statout/" \
            "../$STAT_DIR/$SERVICE/" || {
            log_error "Failed to collect stats from $SERVICE"
            continue
        }
        log_info "Successfully collected stats from $SERVICE"
    else
        log_error "Stats directory not found on $SERVICE"
    fi
done

cd ..
log_info "Data collection completed. Stats saved in $STAT_DIR"