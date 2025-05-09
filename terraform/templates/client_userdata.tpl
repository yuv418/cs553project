#!/bin/bash
set -e

# Update system and install required packages
yum update -y
yum install -y nodejs20 nodejs20-npm git

# Set up environment variables
export HOME=/root
export XDG_CONFIG_HOME=/root/.config
mkdir -p $XDG_CONFIG_HOME

# Create app directory
mkdir -p /opt/flappygo-client
cd /opt/flappygo-client

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
mkdir -p /opt/flappygo-client/certs

# Install provided certificates
echo "Installing certificates..."
echo "${certificate_content}" | base64 -d > /opt/flappygo-client/certs/cert.pem
echo "${private_key_content}" | base64 -d > /opt/flappygo-client/certs/key.pem

# Write SPKI hash for Chrome
echo "${spki_hash}" > /opt/flappygo-client/certs/spki_hash.txt

mkdir -p /opt/flappygo-client/flap-client/dist

# Build the client
cd /opt/flappygo-client/flap-client
npm install

# Create .env file with service URLs
cat > /opt/flappygo-client/flap-client/.env << EOF
VITE_AUTH_SERVICE_URL=https://${auth_url}:${auth_port}
VITE_INITIATOR_SERVICE_URL=https://${initiator_url}:${initiator_port}
VITE_SCORE_SERVICE_URL=https://${score_url}:${score_port}
VITE_WEBTRANSPORT_GAME_URL=https://${engine_url}:4433/gameEngine/GameSession
VITE_WEBTRANSPORT_MUSIC_URL=https://${music_url}:4433/music/MusicSession
VITE_LOG_LATENCY=1
VITE_DEBUG=
EOF

# Build the production version
npm run build

# Install and configure nginx
yum install -y nginx
systemctl enable nginx

# Create custom nginx config
cat > /etc/nginx/conf.d/default.conf << EOF
server {
    listen 80 default_server;
    listen [::]:80 default_server;
    listen 443 ssl default_server;
    listen [::]:443 ssl default_server;
    server_name _;
    
    # SSL configuration
    ssl_certificate /opt/flappygo-client/certs/cert.pem;
    ssl_certificate_key /opt/flappygo-client/certs/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;
    ssl_session_tickets off;
    
    # Force HTTPS except for health check
    if (dollarsignscheme = http) {
        set dollarsignredirect_https 1;
    }
    if (dollarsignrequest_uri = /health) {
        set dollarsignredirect_https 0;
    }
    if (dollarsignredirect_https = 1) {
        return 301 https://dollarsignhostdollarsignrequest_uri;
    }

    location /health {
        return 200 'OK';
    }

    location /spki_hash.txt {
        alias /opt/flappygo-client/certs/spki_hash.txt;
        default_type text/plain;
    }
    
    location / {
        root /opt/flappygo-client/flap-client/dist;
        index index.html;
        try_files dollarsignuri dollarsignuri/ /index.html;
    }
}
EOF

# Replace "dollarsign" in the nginx config with a dollar sign character (I can't figure out why)
sed -i 's/dollarsign/\$/g' /etc/nginx/conf.d/default.conf

# Add instructions for Chrome
cat > /opt/flappygo-client/flap-client/dist/quic-instructions.html << EOF
<!DOCTYPE html>
<html>
<head>
    <title>WebTransport Instructions</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 20px auto; line-height: 1.6; }
        pre { background: #f4f4f4; padding: 10px; overflow-x: auto; }
        .highlight { background: #ffff00; padding: 2px 5px; }
    </style>
</head>
<body>
    <h1>WebTransport Connection Instructions</h1>
    <p>FlappyGo! uses WebTransport which requires special browser flags to work with self-signed certificates.</p>
    
    <h2>Chrome Instructions</h2>
    <p>Close all Chrome instances and restart with the following command:</p>
    <pre id="chrome-command"></pre>
    
    <script>
        fetch('/spki_hash.txt')
            .then(response => response.text())
            .then(hash => {
                const gameUrl = new URL(window.location.href).hostname;
                document.getElementById('chrome-command').textContent = 
                    "chrome --origin-to-force-quic-on=" + gameUrl + ":4433," + gameUrl + ":4434 --ignore-certificate-errors-spki-list=" + hash.trim();
            });
    </script>
</body>
</html>
EOF

# Start nginx
systemctl restart nginx

# Output completion message
echo "FlappyGo client deployment completed successfully!"
