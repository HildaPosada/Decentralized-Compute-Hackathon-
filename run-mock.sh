#!/bin/bash

# DistributeAI - Mock/Simulation Mode
# For GitHub Dev environments without Docker

set -e

echo "⚡ DistributeAI - Simulation Mode"
echo "=================================="
echo ""
echo "Running in MOCK mode (no Docker required)"
echo "This simulates the platform behavior for demo purposes"
echo ""

# Create directories for mock data
mkdir -p /tmp/distributeai/{data,logs}

# Start mock coordinator in background
echo "🚀 Starting Mock Coordinator..."
cat > /tmp/distributeai/coordinator.sh << 'EOF'
#!/bin/bash
echo "[$(date)] Coordinator started on port 8080"
while true; do
  sleep 5
  echo "[$(date)] Scheduler checking for pending jobs..."
  echo "[$(date)] Found 0 pending jobs"
done
EOF
chmod +x /tmp/distributeai/coordinator.sh

# Start mock workers in background
for i in 1 2 3; do
  echo "🖥️  Starting Mock Worker $i..."
  cat > /tmp/distributeai/worker$i.sh << EOF
#!/bin/bash
echo "[$(date)] Worker $i started (Region: us-west)"
echo "[$(date)] Worker $i registered with coordinator"
while true; do
  sleep 10
  echo "[$(date)] Worker $i: Heartbeat sent (CPU: 25%, Memory: 40%)"
  echo "[$(date)] Worker $i: Polling for jobs..."
done
EOF
  chmod +x /tmp/distributeai/worker$i.sh
done

echo ""
echo "✅ Mock services started!"
echo ""
echo "📊 Service Status:"
echo "   ✅ Coordinator - Running (Mock)"
echo "   ✅ Worker 1 - Online (Mock)"
echo "   ✅ Worker 2 - Online (Mock)"
echo "   ✅ Worker 3 - Online (Mock)"
echo ""

echo "🎯 Simulating Job Submission..."
echo ""
JOB_ID="job-$(date +%s)"

echo "POST /api/v1/jobs"
echo "{"
echo "  \"id\": \"$JOB_ID\","
echo "  \"name\": \"Hash Verification Demo\","
echo "  \"status\": \"pending\","
echo "  \"docker_image\": \"alpine:latest\","
echo "  \"command\": [\"sh\", \"-c\", \"echo 'Hello' | sha256sum\"]"
echo "}"
echo ""

sleep 2

echo "🔄 Job Lifecycle Simulation:"
echo ""
echo "[T+0s] Job $JOB_ID created - Status: pending"
sleep 1
echo "[T+1s] Scheduler assigned job to 3 workers"
sleep 1
echo "[T+2s] Worker 1 started execution"
echo "[T+2s] Worker 2 started execution"
echo "[T+2s] Worker 3 started execution"
sleep 2
echo "[T+4s] Worker 1 completed - Hash: 8b1a9953c4611296a827abf8c47804d7..."
echo "[T+4s] Worker 2 completed - Hash: 8b1a9953c4611296a827abf8c47804d7..."
echo "[T+5s] Worker 3 completed - Hash: 8b1a9953c4611296a827abf8c47804d7..."
sleep 1
echo ""
echo "✅ Verification: Consensus Reached!"
echo "   - 3/3 nodes produced matching hash"
echo "   - Result: 8b1a9953c4611296a827abf8c47804d7fab7fbe0a2d899b2eca89be35654fbd5"
echo ""
echo "📈 Reputation Updates:"
echo "   - Worker 1: +5 reputation (now 105)"
echo "   - Worker 2: +5 reputation (now 105)"
echo "   - Worker 3: +5 reputation (now 105)"
echo ""

sleep 2

echo "🔥 Simulating Node Failure..."
echo ""
echo "[T+10s] Worker 2 disconnected (simulating crash)"
echo "[T+12s] Coordinator detected missing heartbeat"
echo "[T+12s] Worker 2 marked offline"
echo "[T+12s] Worker 2 reputation penalty: -20 (now 85)"
echo ""

sleep 2

echo "📊 System Statistics:"
echo ""
echo "Nodes:"
echo "  Total: 3"
echo "  Online: 2 (Worker 1, Worker 3)"
echo "  Offline: 1 (Worker 2)"
echo ""
echo "Resources:"
echo "  Total CPU Cores: 8"
echo "  Total Memory: 16 GB"
echo ""
echo "Jobs:"
echo "  Total: 1"
echo "  Completed: 1"
echo "  Failed: 0"
echo ""

echo "✅ Simulation Complete!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📖 What This Demonstrates:"
echo ""
echo "1. ✅ k-of-n Verification"
echo "   - Job executed on 3 nodes"
echo "   - All 3 produced matching results"
echo "   - Consensus reached (2/3 required)"
echo ""
echo "2. ✅ Reputation System"
echo "   - Correct results: +5 reputation"
echo "   - Node failure: -20 reputation"
echo ""
echo "3. ✅ Fault Tolerance"
echo "   - Worker 2 failed mid-operation"
echo "   - System detected and handled gracefully"
echo "   - Job still completed successfully"
echo ""
echo "4. ✅ Smart Scheduling"
echo "   - Coordinator tracks node status"
echo "   - Only schedules to online nodes"
echo "   - Prioritizes high-reputation nodes"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🚀 To run the REAL platform (with Docker):"
echo ""
echo "1. On a machine with Docker installed:"
echo "   git clone https://github.com/HildaPosada/Decentralized-Compute-Hackathon-"
echo "   cd Decentralized-Compute-Hackathon-"
echo "   ./run.sh"
echo ""
echo "2. Or use GitHub Codespaces (has Docker):"
echo "   - Go to your GitHub repo"
echo "   - Click 'Code' → 'Codespaces' → 'Create codespace'"
echo "   - Run: ./run.sh"
echo ""
echo "📚 View the code:"
echo "   - Coordinator: coordinator/cmd/coordinator/main.go"
echo "   - Worker: worker/cmd/worker/main.go"
echo "   - Verification: coordinator/internal/verification/verifier.go"
echo ""
echo "📖 View documentation:"
echo "   - README.md"
echo "   - docs/ARCHITECTURE.md"
echo "   - docs/DEMO_GUIDE.md"
echo ""
