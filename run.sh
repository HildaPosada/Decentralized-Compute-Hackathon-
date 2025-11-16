#!/bin/bash

# DistributeAI - One-Command Startup Script
# For Decentralized Compute Challenge Hackathon

set -e

echo "🚀 DistributeAI - Decentralized Compute Network"
echo "================================================"
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file from .env.example..."
    cp .env.example .env
fi

# Build and start all services
echo "🏗️  Building containers..."
docker-compose build

echo ""
echo "▶️  Starting services..."
docker-compose up -d

echo ""
echo "⏳ Waiting for services to be healthy..."
sleep 10

# Check service health
echo ""
echo "🏥 Checking service health..."
docker-compose ps

echo ""
echo "✅ DistributeAI is now running!"
echo ""
echo "📊 Access Points:"
echo "   - Coordinator API:  http://localhost:8080"
echo "   - Dashboard UI:     http://localhost:3000"
echo "   - Prometheus:       http://localhost:9090"
echo "   - Grafana:          http://localhost:3001 (admin/admin)"
echo "   - MinIO Console:    http://localhost:9001 (minioadmin/minioadmin)"
echo ""
echo "🔧 Quick Commands:"
echo "   - Submit a job:     ./cli/bin/distributeai submit examples/hash-verify/job.json"
echo "   - View logs:        docker-compose logs -f coordinator"
echo "   - Stop services:    docker-compose down"
echo ""
echo "📖 For more information, see README.md"
