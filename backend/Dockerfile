FROM golang:1.24 AS builder
ARG SERVICE

# Set working directory
WORKDIR /app

# Install Protocol Buffers
ARG PROTOC_VERSION=30.2
RUN apt-get update && apt-get install -y unzip curl && \
    curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip && \
    unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip -d /usr/local && \
    rm protoc-${PROTOC_VERSION}-linux-x86_64.zip && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Install Go protobuf plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Copy source code
COPY backend/ ./

# Download dependencies
RUN go mod download

# Ensure PATH includes protoc and Go binaries
ENV PATH="$PATH:/usr/local/bin:$GOPATH/bin"

# Generate protobufs and build the application
RUN make $SERVICE

# To create a smaller final image
FROM ubuntu:24.04
ARG SERVICE

RUN apt-get update && \
    apt-get install -y ca-certificates curl unzip && \
    update-ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Install Protocol Buffers runtime in the final image (needed for runtime)
ARG PROTOC_VERSION=30.2
RUN curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip && \
    mkdir -p /usr/local && \
    unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip -d /usr/local && \
    rm protoc-${PROTOC_VERSION}-linux-x86_64.zip

# Add protoc to PATH
ENV PATH="$PATH:/usr/local/bin"

# Configure UDP buffer sizes for WebTransport
RUN echo "net.core.rmem_max=7500000" >> /etc/sysctl.d/99-webtransport.conf && \
    echo "net.core.wmem_max=7500000" >> /etc/sysctl.d/99-webtransport.conf

# Expose gRPC and WebTransport ports
EXPOSE 50051-50056/tcp
EXPOSE 4433/tcp
EXPOSE 4433/udp

# Copy the binary from the builder stage
WORKDIR /app
COPY --from=builder /app/out/${SERVICE} /app/

# Copy TLS certificates
COPY certs/cert.pem /app/cert.pem
COPY certs/key.pem /app/key.pem

# Set environment variables
ENV MICROSERVICE=1

# Expose ports
EXPOSE 50051-50059
EXPOSE 4433

# Use tini as init to properly handle signals
RUN apt-get update && apt-get install -y tini && rm -rf /var/lib/apt/lists/*
ENTRYPOINT ["/usr/bin/tini", "--"]

# Run the binary with the service name
ARG SERVICE=monolith
ENV SERVICE_BIN=$SERVICE
CMD ["/bin/bash", "-c", "/app/$SERVICE_BIN"]
