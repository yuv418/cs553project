#!/usr/bin/env bash
set -e

# Helper functions
function print_usage() {
  echo "Usage: $0 [OPTIONS]"
  echo "Options:"
  echo "  --mode MODE                Set deployment mode (monolith|microservices) [default: monolith]"
  echo "  --generate-certs           Generate new certificates"
  echo "  --chrome-flags             Print Chrome flags for WebTransport"
  echo "  --help                     Print this help message"
}

# Default values
DEPLOYMENT_MODE="monolith"
GENERATE_CERTS=false
PRINT_CHROME_FLAGS=false

# Parse arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    --mode)
      DEPLOYMENT_MODE="$2"
      shift 2
      ;;
    --generate-certs)
      GENERATE_CERTS=true
      shift
      ;;
    --chrome-flags)
      PRINT_CHROME_FLAGS=true
      shift
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

# Create certs directory if it doesn't exist
mkdir -p certs

# Generate certificates if requested or if they don't exist
if [[ "$GENERATE_CERTS" = true ]] || [[ ! -f certs/cert.pem ]] || [[ ! -f certs/key.pem ]]; then
  echo "Generating TLS certificates..."
  ./generate_certs.sh --output-dir ./certs
fi

# Extract SPKI hash if not already saved
if [[ ! -f certs/spki_hash.txt ]] || [[ "$GENERATE_CERTS" = true ]]; then
  SPKI_HASH=$(openssl x509 -in certs/cert.pem -pubkey -noout | \
              openssl pkey -pubin -outform der | \
              openssl dgst -sha256 -binary | \
              base64)
  echo "$SPKI_HASH" > certs/spki_hash.txt
else
  SPKI_HASH=$(cat certs/spki_hash.txt)
fi

# Print Chrome flags if requested
if [[ "$PRINT_CHROME_FLAGS" = true ]]; then
  echo "To use WebTransport with Chrome, run:"
  echo "-------------------------------------------------------------------------"
  echo "chrome --origin-to-force-quic-on=localhost:4433,localhost:4434 --ignore-certificate-errors-spki-list=$SPKI_HASH"
  echo "-------------------------------------------------------------------------"
  exit 0
fi

# Stop any existing containers
echo "Stopping any existing FlappyGo! containers..."
docker-compose down || true

# Start the application
echo "Starting FlappyGo! in $DEPLOYMENT_MODE mode..."
export DEPLOYMENT_MODE=$DEPLOYMENT_MODE
docker-compose --profile $DEPLOYMENT_MODE up --build -d

# Print WebTransport URLs
echo "-------------------------------------------------------------------------"
echo "FlappyGo! is now running!"
echo
echo "Client URL: http://localhost:8080"
echo
if [[ "$DEPLOYMENT_MODE" = "monolith" ]]; then
  echo "WebTransport URLs:"
  echo "Game:  https://localhost:4433/gameEngine/GameSession"
  echo "Music: https://localhost:4433/music/MusicSession"
else
  echo "WebTransport URLs:"
  echo "Game:  https://localhost:4433/gameEngine/GameSession"
  echo "Music: https://localhost:4434/music/MusicSession"
fi
echo
echo "Your browser must be started with special flags!"
echo "-------------------------------------------------------------------------"
echo "chrome --origin-to-force-quic-on=localhost:4433,localhost:4434 --ignore-certificate-errors-spki-list=$SPKI_HASH"
echo "-------------------------------------------------------------------------"
