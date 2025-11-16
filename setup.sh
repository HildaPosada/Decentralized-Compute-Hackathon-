#!/bin/bash

# Setup script to initialize Go modules
set -e

echo "🔧 Setting up DistributeAI..."
echo ""

# Fix CLI module
echo "📦 Initializing CLI module..."
cd cli
go mod tidy
echo "✅ CLI module ready"
echo ""

# Fix Coordinator module
cd ../coordinator
echo "📦 Initializing Coordinator module..."
go mod tidy
echo "✅ Coordinator module ready"
echo ""

# Fix Worker module
cd ../worker
echo "📦 Initializing Worker module..."
go mod tidy
echo "✅ Worker module ready"
echo ""

cd ..

echo "✅ All modules initialized!"
echo ""
echo "You can now:"
echo "  1. Start the platform: ./run.sh"
echo "  2. Use the CLI: ./bin/distributeai --help"
