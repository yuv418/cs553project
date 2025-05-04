#!/usr/bin/env bash
set -e

# Helper functions
function print_usage() {
  echo "Usage: $0 [OPTIONS]"
  echo "Options:"
  echo "  --output-dir DIR           Directory to store certificates (default: ./certs)"
  echo "  --common-name NAME         Common Name (CN) for certificate (default: FlappyGo)"
  echo "  --days DAYS                Validity period in days (default: 365)"
  echo "  --country COUNTRY          Country code (default: US)"
  echo "  --state STATE              State (default: California)"
  echo "  --locality LOCALITY        Locality (default: San Francisco)"
  echo "  --organization ORG         Organization (default: FlappyGo)"
  echo "  --help                     Print this help message"
}

# Default values
OUTPUT_DIR="./certs"
COMMON_NAME="FlappyGo"
DAYS=365
COUNTRY="US"
STATE="California"
LOCALITY="San Francisco"
ORGANIZATION="FlappyGo"

# Parse arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    --output-dir)
      OUTPUT_DIR="$2"
      shift 2
      ;;
    --common-name)
      COMMON_NAME="$2"
      shift 2
      ;;
    --days)
      DAYS="$2"
      shift 2
      ;;
    --country)
      COUNTRY="$2"
      shift 2
      ;;
    --state)
      STATE="$2"
      shift 2
      ;;
    --locality)
      LOCALITY="$2"
      shift 2
      ;;
    --organization)
      ORGANIZATION="$2"
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

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Generate OpenSSL configuration file
cat > "$OUTPUT_DIR/openssl.cnf" << EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = $COUNTRY
ST = $STATE
L = $LOCALITY
O = $ORGANIZATION
CN = $COMMON_NAME

[v3_req]
keyUsage = digitalSignature, keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = $COMMON_NAME
DNS.2 = localhost
IP.1 = 127.0.0.1
EOF

# Generate private key and certificate
echo "Generating TLS certificates..."
openssl req -x509 -newkey rsa:2048 \
  -keyout "$OUTPUT_DIR/key.pem" \
  -out "$OUTPUT_DIR/cert.pem" \
  -days $DAYS \
  -nodes \
  -config "$OUTPUT_DIR/openssl.cnf" \
  -extensions v3_req

# Calculate fingerprint for Chrome's --ignore-certificate-errors-spki-list flag
SPKI_HASH=$(openssl x509 -in "$OUTPUT_DIR/cert.pem" -pubkey -noout | \
            openssl pkey -pubin -outform der | \
            openssl dgst -sha256 -binary | \
            base64)

# Output success message
echo "Certificate generation complete!"
echo "---------------------------------"
echo "Certificate: $OUTPUT_DIR/cert.pem"
echo "Private key: $OUTPUT_DIR/key.pem"
echo "---------------------------------"
echo "Certificate validity: $DAYS days"
echo "For Chrome, use the following flag to trust this certificate:"
echo "--ignore-certificate-errors-spki-list=$SPKI_HASH"
echo "---------------------------------"
echo "Example command to use with chrome:"
echo "chrome --origin-to-force-quic-on=localhost:4433 --ignore-certificate-errors-spki-list=$SPKI_HASH"
echo "---------------------------------"

# Write the fingerprint to a file for ease
echo "$SPKI_HASH" > "$OUTPUT_DIR/spki_hash.txt"

# Clean up temporary files
rm "$OUTPUT_DIR/openssl.cnf"
