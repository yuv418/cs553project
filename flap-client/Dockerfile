FROM node:20 AS builder

# Set working directory
WORKDIR /app

# Copy package.json and package-lock.json
COPY flap-client/package*.json ./

# Install dependencies
RUN npm ci

# Copy protos
COPY backend/protos ../backend/protos

# Copy the rest of the code
COPY flap-client/ ./

# Generate protobuf files
RUN npx buf generate

# Build the app - use the deployment mode to select the right env file
ARG DEPLOYMENT_MODE
RUN if [ "$DEPLOYMENT_MODE" = "microservices" ]; then \
    cp .env.microservices .env; \
    else \
    cp .env.monolith .env; \
    fi
RUN npm run build

# Start with fresh node for runtime
FROM node:20-alpine

# Install serve for static file hosting
RUN npm install -g serve

# Set working directory
WORKDIR /app

# Copy built assets from the builder stage
COPY --from=builder /app/dist /app

# Expose port
EXPOSE 8080

# Serve
CMD ["serve", "-s", ".", "-l", "8080"]
