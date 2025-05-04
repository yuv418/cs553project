# FlappyGo!

FlappyGo! is an enterprise-ready distribution of the classic game "Flappy Bird". It supports an optional, distributed architecture for its components. Deployers can choose to deploy FlappyGo! as a monolith, with its constituent components running in a single serverside app (plus client), or as microservices, with different components running in different locations and communicating.

## Overview

The `backend` directory contains the game server, which is written in Go. It supports a variety of deployment patterns, where application binaries can be built via `make`.

## Deployment Options

FlappyGo! can be deployed in various configurations to support research on latency and performance across different architectures:

1. **Monolith** - All services running on a single machine
2. **Microservices on Single Machine** - Services split but running on the same host
3. **Microservices Across Availability Zones** - Services distributed across multiple AZs in one region
4. **Microservices Across Regions** - Services distributed across multiple regions
5. **Microservices Globally Distributed** - Services deployed globally for maximum distribution

These deployment patterns can be configured using our Terraform infrastructure as code.

## Setup Instructions

### Prerequisites

- [Go](https://go.dev/doc/install) 1.24+ 
- [Protocol Buffers](https://protobuf.dev/installation/) 30+
- [Terraform](https://www.terraform.io/downloads.html) v1.2.0+ (for cloud deployment)
- [AWS CLI](https://aws.amazon.com/cli/) configured with appropriate credentials (for cloud deployment)
- [Docker](https://docs.docker.com/get-docker/) (for containerized deployment)
- [Node.js and npm](https://nodejs.org/) (for client development)

### System Configuration

WebTransport requires adequate UDP buffer sizes:

```bash
# Apply temporarily
sysctl -w net.core.rmem_max=7500000
sysctl -w net.core.wmem_max=7500000

# Make permanent
echo "net.core.rmem_max=7500000" >> /etc/sysctl.conf
echo "net.core.wmem_max=7500000" >> /etc/sysctl.conf
sysctl -p
```

## Quick Start

To quickly run FlappyGo! with proper WebTransport support:

```bash
# Set up certificates and generate browser flags
./run.sh --generate-certs

# Run in monolith mode (default)
./run.sh

# Or run in microservices mode
./run.sh --mode microservices
```

## Detailed Deployment Instructions

### Monolithic Deployment

Deploying FlappyGo! monolithically can be accomplished by executing:

```bash
# Build the game
make monolith

# Run the game
MICROSERVICE=0 \
AUTH_URL=localhost:50051 \
WORLD_GEN_URL=localhost:50051 \
INITIATOR_URL=localhost:50051 \
GAME_ENGINE_URL=localhost:50051 \
MUSIC_URL=localhost:50051 \
SCORE_URL=localhost:50051 \
./out/monolith
```

### Microservice-based Deployment

Deploying FlappyGo! as microservices:

```bash
# Build specific component
make <component>  # where <component> is one of: initiator, worldgen, engine, auth, music, or score

# Run component
./out/<component> --addr=localhost:50054  # replace port as needed
```

NOTE: Different components require communication with specific other components. Pass the URLs of those services as environment variables when executing the binaries (see Monolith run command above).

#### Docker Local Deployment

For local testing, you can use Docker Compose:

```bash
# Run as monolith
docker-compose --profile monolith up

# Run as microservices
docker-compose --profile microservices up
```

## Client Setup

The `flap-client` directory contains the game client, which is written in TypeScript and accessed through a browser.

### Setup

```bash
# Install dependencies
cd flap-client
npm install

# Generate protobuf files
npx buf generate

# Create or update .env file from sample
cp .env.sample .env
# Edit .env to point to your services
```

### Running

```bash
# Development with auto-reload
npm run dev

# Build for production
npm run build
# Output will be in the dist directory
```

## WebTransport and HTTP/3 Support

FlappyGo! uses WebTransport for real-time communication between the client and server components, providing low-latency bidirectional streams essential for game performance.

### Browser Configuration

To use WebTransport with FlappyGo!, Chrome requires special flags because of the self-signed certificates:

1. Close all Chrome instances
2. Run Chrome with the following flags:

```bash
chrome --origin-to-force-quic-on=localhost:4433,localhost:4434 --ignore-certificate-errors-spki-list=YOUR_SPKI_HASH
```

Where `YOUR_SPKI_HASH` is the base64-encoded SHA-256 hash of the certificate's public key.

To get the hash, run:

```bash
openssl x509 -in certs/cert.pem -pubkey -noout | \
openssl pkey -pubin -outform der | \
openssl dgst -sha256 -binary | \
base64
```

Our `run.sh` script can also print these flags for you:

```bash
./run.sh --chrome-flags
```

## TLS Certificates

FlappyGo! requires TLS certificates for secure communication. The application supports both self-signed certificates for development and custom certificates for production.

### Development Certificates

By default, the application looks for certificates in the following locations:
- `cert.pem` - TLS certificate
- `key.pem` - TLS private key

For local development, you can use the certificates provided in the `certs` directory or generate new ones with our helper script:

```bash
./generate_certs.sh
```

This will create a new set of certificates in the `./certs` directory and provide instructions for configuring your browser to trust them.

## Credits

### Asset Credit

Select assets (sprites, audio) courtesy [Samuel Custodio](https://github.com/samuelcust/flappy-bird-assets).

### Original Reference Game AI Attribution

An initial `reference_game.html` was generated via OpenAI's API using `o1` model with the following prompt:

> Generate HTML/CSS/JS for a clone of the game Flappy Bird. Ensure complete feature parity with the original Flappy Bird game originally released for mobile devices. Output the list of features first, then write the code. Use basic inline SVGs for temporary graphics.

The model returned a detailed Flappy Bird implementation with the following features:

- Side-Scrolling Environment with continuous, seamless looping background
- Gravity and Flap Mechanics with realistic physics
- Pipes with Random Gaps procedurally generated
- Collision Detection with pipes and ground
- Score Tracking and high score storage
- Difficulty Progression as game advances
- Animated Ground with scrolling effect
- Multiple Game States (waiting, playing, game over)
- Sound Effects for actions and events
- Tap/Click Controls for intuitive play
- Game Over Screen with score display and restart option
