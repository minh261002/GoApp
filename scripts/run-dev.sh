#!/bin/bash

# Development run script - Only run main app server
# This script runs only the main application without background workers

echo "🚀 Starting Go App in Development Mode (App Server Only)"
echo "=================================================="

# Set environment variables for development
export LOG_LEVEL=warn
export GIN_MODE=debug
export PORT=8080
export HOST=0.0.0.0

# Database configuration
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=go_app_user
export DB_PASSWORD=go_app_password
export DB_NAME=go_app_db

# Redis configuration
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=
export REDIS_DB=0

# JWT configuration
export JWT_SECRET=your-jwt-secret-key-change-in-production
export JWT_ACCESS_EXPIRY=24
export JWT_REFRESH_EXPIRY=168

echo "📋 Environment Configuration:"
echo "  - LOG_LEVEL: $LOG_LEVEL"
echo "  - GIN_MODE: $GIN_MODE"
echo "  - PORT: $PORT"
echo "  - DB_HOST: $DB_HOST"
echo "  - REDIS_HOST: $REDIS_HOST"
echo ""

echo "🔧 Building application..."
go build -o bin/server ./cmd/server

if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo "✅ Build successful!"
echo ""

echo "🌐 Starting server on http://localhost:$PORT"
echo "   Press Ctrl+C to stop"
echo ""

# Run the application
./bin/server

