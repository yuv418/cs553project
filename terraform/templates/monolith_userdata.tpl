#!/bin/bash
set -e

# Update system and install required packages
yum update -y
yum install -y golang git make protobuf-compiler protobuf-devel openssl

# Set up Go environment
export HOME=/root
export GOPATH=/root/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
export PROTOC_INCLUDE=/usr/include
mkdir -p $GOPATH/bin

# Configure system for WebTransport UDP buffer sizes
echo "net.core.rmem_max=7500000" >> /etc/sysctl.conf
echo "net.core.wmem_max=7500000" >> /etc/sysctl.conf
sysctl -p

# Create app directory
mkdir -p /opt/flappygo
cd /opt/flappygo

# Clone repository (with token if provided)
if [ -n "${github_token}" ]; then
    # Disable command printing to avoid exposing token in logs
    set +x
    git clone https://${github_token}@github.com/yuv418/cs553project . >/dev/null 2>&1
    set -x
else
    git clone https://github.com/yuv418/cs553project .
fi

# Create TLS certificates directory
mkdir -p /opt/flappygo/certs

# Install provided certificates
echo "Installing certificates..."
echo "${certificate_content}" | base64 -d > /opt/flappygo/certs/cert.pem
echo "${private_key_content}" | base64 -d > /opt/flappygo/certs/key.pem

# Write SPKI hash for Chrome
echo "${spki_hash}" > /opt/flappygo/certs/spki_hash.txt

# Install Go tools
go env -w GO111MODULE=on
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Build the monolith
cd /opt/flappygo/backend
make monolith

# Create systemd service file
cat > /etc/systemd/system/flappygo.service << EOF
[Unit]
Description=FlappyGo Monolith Service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/flappygo/backend
Environment="MICROSERVICE=0"
Environment="AUTH_URL=localhost:${service_ports.auth}"
Environment="WORLD_GEN_URL=localhost:${service_ports.worldgen}"
Environment="INITIATOR_URL=localhost:${service_ports.initiator}"
Environment="GAME_ENGINE_URL=localhost:${service_ports.engine}"
Environment="MUSIC_URL=localhost:${service_ports.music}"
Environment="SCORE_URL=localhost:${service_ports.score}"
Environment="AUTH_CERT_FILE=/opt/flappygo/certs/cert.pem"
Environment="AUTH_KEY_FILE=/opt/flappygo/certs/key.pem"
Environment="PORT_RANGE_START=50051"
Environment="PORT_RANGE_END=50059"
Environment="WEBTRANSPORT_PORT=4433"
ExecStart=/opt/flappygo/backend/out/monolith
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Enable and start the service
systemctl daemon-reload
systemctl enable flappygo
systemctl start flappygo

echo "FlappyGo monolith deployment completed successfully!"