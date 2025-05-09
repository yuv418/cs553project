#!/usr/bin/env bash
set -e

# Helper functions
function print_usage() {
  echo "Usage: $0 [OPTIONS]"
  echo "Options:"
  echo "  --deployment-mode MODE     Set deployment mode (monolith|microservices)"
  echo "  --deployment-pattern PATTERN   Set deployment pattern (single_instance|multi_az|multi_region)"
  echo "  --aws-region REGION        Set AWS region"
  echo "  --instance-type TYPE       Set EC2 instance type"
  echo "  --key-name KEY             Set SSH key name for EC2 instances"
  echo "  --certificate-path PATH    Path to custom TLS certificate (optional)"
  echo "  --private-key-path PATH    Path to custom TLS private key (optional)"
  echo "  --github-token TOKEN       GitHub Personal Access Token for private repository access"
  echo "  --ssh-key-path PATH        Path to SSH private key for instance access"
  echo "  --help                     Print this help message"
}

function generate_ssh_key() {
    local key_dir="$1"
    local key_name="$2"
    mkdir -p "$key_dir"
    ssh-keygen -t rsa -b 4096 -f "${key_dir}/${key_name}" -N "" -C "flappygo-deployment-key"
    chmod 600 "${key_dir}/${key_name}"
    chmod 644 "${key_dir}/${key_name}.pub"
    echo "${key_dir}/${key_name}"
}

function check_existing_ssh_key() {
    local key_path="$(dirname "$0")/certs/ssh_key"
    if [[ -f "$key_path" ]]; then
        echo "$key_path"
        return 0
    fi
    return 1
}

# Default values
DEPLOYMENT_MODE="monolith"
DEPLOYMENT_PATTERN="single_instance"
AWS_REGION="us-east-1"
USE_LOAD_BALANCER=false
INSTANCE_TYPE="t2.micro"
KEY_NAME=""
DOMAIN_NAME=""
CERTIFICATE_PATH=""
PRIVATE_KEY_PATH=""
USE_OWN_CERTIFICATES=false
GITHUB_TOKEN=""
SSH_KEY_PATH=""

# Parse arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    --deployment-mode)
      DEPLOYMENT_MODE="$2"
      shift 2
      ;;
    --deployment-pattern)
      DEPLOYMENT_PATTERN="$2"
      shift 2
      ;;
    --aws-region)
      AWS_REGION="$2"
      shift 2
      ;;
    --instance-type)
      INSTANCE_TYPE="$2"
      shift 2
      ;;
    --key-name)
      KEY_NAME="$2"
      shift 2
      ;;
    --certificate-path)
      CERTIFICATE_PATH="$2"
      USE_OWN_CERTIFICATES=true
      shift 2
      ;;
    --private-key-path)
      PRIVATE_KEY_PATH="$2"
      USE_OWN_CERTIFICATES=true
      shift 2
      ;;
    --github-token)
      GITHUB_TOKEN="$2"
      shift 2
      ;;
    --ssh-key-path)
      SSH_KEY_PATH="$2"
      shift 2
      ;;
    --help)
      print_usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      print_usage
      exit 1
      ;;
  esac
done

# Validate deployment mode
if [[ "$DEPLOYMENT_MODE" != "monolith" && "$DEPLOYMENT_MODE" != "microservices" ]]; then
  echo "Error: Deployment mode must be 'monolith' or 'microservices'"
  exit 1
fi

# Validate deployment pattern
if [[ "$DEPLOYMENT_PATTERN" != "single_instance" && "$DEPLOYMENT_PATTERN" != "multi_az" && "$DEPLOYMENT_PATTERN" != "multi_region" ]]; then
  echo "Error: Deployment pattern must be 'single_instance', 'multi_az', or 'multi_region'"
  exit 1
fi

# Check terraform is installed
if ! command -v terraform &> /dev/null; then
  echo "Error: Terraform is not installed"
  exit 1
fi

# Initialize terraform
cd "$(dirname "$0")/terraform"
terraform init

# Check for existing SSH key or generate new one
if [[ -z "$SSH_KEY_PATH" ]]; then
    if FOUND_KEY=$(check_existing_ssh_key); then
        echo "Found existing SSH key at: $FOUND_KEY"
        SSH_KEY_PATH="$FOUND_KEY"
    else
        echo "No SSH key provided or found, generating one..."
        CERTS_DIR="$(dirname "$0")/certs"
        mkdir -p "$CERTS_DIR"
        SSH_KEY_PATH=$(generate_ssh_key "$CERTS_DIR" "ssh_key")
    fi
    KEY_NAME="flappygo-key2"
fi

# Validate SSH key
if [[ ! -f "$SSH_KEY_PATH" ]]; then
    echo "Error: SSH key file not found: $SSH_KEY_PATH"
    exit 1
fi

if [[ "$(stat -c %a "$SSH_KEY_PATH")" != "600" ]]; then
    echo "Warning: SSH key file permissions should be 600. Fixing..."
    chmod 600 "$SSH_KEY_PATH"
fi

# Generate tfvars file
cat > terraform.tfvars << EOF
deployment_mode = "$DEPLOYMENT_MODE"
deployment_pattern = "$DEPLOYMENT_PATTERN"
aws_region = "$AWS_REGION"
instance_type = "$INSTANCE_TYPE"
use_own_certificates = $USE_OWN_CERTIFICATES
github_token = "${GITHUB_TOKEN}"
ssh_private_key_path = "${SSH_KEY_PATH}"
EOF

# Add key name if provided
if [[ -n "$KEY_NAME" ]]; then
  echo "key_name = \"$KEY_NAME\"" >> terraform.tfvars
fi

# Add certificate paths if provided
if [[ -n "$CERTIFICATE_PATH" ]]; then
  echo "certificate_path = \"$CERTIFICATE_PATH\"" >> terraform.tfvars
fi

# Add certificate path if provided
if [[ -n "$CERTIFICATE_PATH" ]]; then
  echo "certificate_path = \"$CERTIFICATE_PATH\"" >> terraform.tfvars
fi

# Add private key path if provided
if [[ -n "$PRIVATE_KEY_PATH" ]]; then
  echo "private_key_path = \"$PRIVATE_KEY_PATH\"" >> terraform.tfvars
fi

# Print deployment summary
echo "Deploying FlappyGo! with the following configuration:"
echo "  Deployment Mode: $DEPLOYMENT_MODE"
echo "  Deployment Pattern: $DEPLOYMENT_PATTERN"
echo "  AWS Region: $AWS_REGION"
echo "  Instance Type: $INSTANCE_TYPE"
if [[ -n "$KEY_NAME" ]]; then
  echo "  SSH Key Name: $KEY_NAME"
fi
if [[ "$USE_OWN_CERTIFICATES" == true ]]; then
  echo "  Using Custom Certificates: Yes"
  echo "  Certificate Path: $CERTIFICATE_PATH"
  echo "  Private Key Path: $PRIVATE_KEY_PATH"
else
  echo "  Using Custom Certificates: No (will generate self-signed)"
fi
if [[ -n "$GITHUB_TOKEN" ]]; then
  echo "  Using GitHub Token: Yes"
fi
if [[ -n "$CERTIFICATE_PATH" ]]; then
  echo "  Certificate Path: $CERTIFICATE_PATH"
fi
if [[ -n "$PRIVATE_KEY_PATH" ]]; then
  echo "  Private Key Path: $PRIVATE_KEY_PATH"
fi
if [[ -n "$SSH_KEY_PATH" ]]; then
    if [[ "$SSH_KEY_PATH" == *"/certs/ssh_key" ]]; then
        echo "  SSH Key: Auto-generated (saved to $SSH_KEY_PATH)"
    else
        echo "  SSH Key Path: $SSH_KEY_PATH"
    fi
fi

# Confirm deployment
read -p "Proceed with deployment? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "Deployment cancelled."
  exit 0
fi

# Run terraform plan
terraform plan -var-file=terraform.tfvars -out=tfplan

# Confirm terraform plan
read -p "Apply the above plan? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "Deployment cancelled."
  exit 0
fi

# Write the deploy time to a file
DEPLOY_TIME_FILE="deploy_time.txt"
echo $(date +"%Y%m%d_%H%M%S") > "$DEPLOY_TIME_FILE"
echo "Deployment time recorded in $DEPLOY_TIME_FILE"

# write deploy type to a file
DEPLOY_TYPE_FILE="deploy_type.txt"
echo "$DEPLOYMENT_MODE_$DEPLOYMENT_PATTERN" > $DEPLOY_TYPE_FILE


# Apply terraform plan
terraform apply tfplan

# Show outputs
echo "============================================================"
echo "Deployment completed successfully!"
echo "Note: It may take a few minutes for the instances to be fully initialized."
echo "Service endpoints:"
terraform output -json service_endpoints

echo "============================================================"
echo "To connect to instances using SSH:"
echo "ssh -i \"$SSH_KEY_PATH\" ec2-user@<instance_public_dns>"
echo "============================================================"
