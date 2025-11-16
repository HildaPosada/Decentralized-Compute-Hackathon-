# 🚀 Deployment Guide - DistributeAI

## Prerequisites

- **Docker** version 20.10+
- **Docker Compose** version 2.0+
- At least 4GB RAM available
- Ports available: 3000, 8080, 9090, 9001, 5432, 6379

---

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/HildaPosada/Decentralized-Compute-Hackathon-
cd Decentralized-Compute-Hackathon-
```

### 2. Start All Services

```bash
./run.sh
```

This will:
- Build all Docker images
- Start PostgreSQL, Redis, MinIO
- Launch Coordinator API
- Start 3 Worker nodes
- Deploy Dashboard UI
- Start Prometheus & Grafana

### 3. Verify Services

Wait 30 seconds for all services to start, then check:

```bash
# Check all containers are running
docker-compose ps

# Test API
curl http://localhost:8080/health

# Test dashboard
curl http://localhost:3000
```

### 4. Access Services

| Service | URL | Credentials |
|---------|-----|-------------|
| **Dashboard** | http://localhost:3000 | N/A |
| **Coordinator API** | http://localhost:8080 | N/A |
| **Prometheus** | http://localhost:9090 | N/A |
| **Grafana** | http://localhost:3001 | admin/admin |
| **MinIO Console** | http://localhost:9001 | minioadmin/minioadmin |

---

## Testing the Platform

### Submit a Test Job

```bash
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Job",
    "docker_image": "alpine:latest",
    "command": ["echo", "Hello DistributeAI"],
    "required_cpu": 1,
    "required_memory": 1
  }'
```

### Check Job Status

```bash
# Save the job ID from above, then:
curl http://localhost:8080/api/v1/jobs/{JOB_ID}
```

### View Nodes

```bash
curl http://localhost:8080/api/v1/nodes
```

### View System Stats

```bash
curl http://localhost:8080/stats
```

---

## Environment Without Docker

If Docker is not available in your current environment, you have these options:

### Option 1: Run Demo Mode

```bash
./demo.sh
```

This shows project statistics and structure without requiring Docker.

### Option 2: Deploy to Cloud

**GitHub Codespaces** (Easiest):
1. Fork the repository on GitHub
2. Open in Codespaces
3. Run `./run.sh`

**AWS EC2**:
```bash
# Launch Ubuntu EC2 instance
# SSH into instance
sudo apt update
sudo apt install -y docker.io docker-compose
sudo usermod -aG docker ubuntu
# Clone repo and run ./run.sh
```

**Azure Container Instances**:
```bash
az container create --resource-group myResourceGroup \
  --name distributeai --image distributeai:latest
```

---

## Building from Source

If you want to build components individually:

### Coordinator
```bash
cd coordinator
go build -o ../bin/coordinator ./cmd/coordinator
```

### Worker
```bash
cd worker
go build -o ../bin/worker ./cmd/worker
```

### CLI
```bash
cd cli
go build -o ../bin/distributeai ./cmd/distributeai
```

### Dashboard
```bash
cd dashboard
npm install
npm run build
```

---

## Configuration

### Environment Variables

Copy `.env.example` to `.env` and customize:

```bash
cp .env.example .env
```

Key variables:
- `POSTGRES_PASSWORD` - Database password
- `MINIO_ROOT_USER` - MinIO username
- `MINIO_ROOT_PASSWORD` - MinIO password
- `VERIFICATION_REDUNDANCY` - Number of nodes per job (default: 3)
- `VERIFICATION_CONSENSUS` - Consensus threshold (default: 2)

### Scaling Workers

To add more worker nodes, edit `docker-compose.yml`:

```yaml
worker4:
  build:
    context: ./worker
    dockerfile: ../deployments/docker/Dockerfile.worker
  environment:
    WORKER_ID: worker-004
    WORKER_NAME: "Worker Node 4"
```

Then restart:
```bash
docker-compose up -d --scale worker=4
```

---

## Stopping the Platform

### Stop all services
```bash
docker-compose stop
```

### Stop and remove containers
```bash
docker-compose down
```

### Stop and remove volumes (clean slate)
```bash
docker-compose down -v
```

---

## Troubleshooting

### Services won't start

**Check Docker is running:**
```bash
docker ps
```

**Check logs:**
```bash
docker-compose logs coordinator
docker-compose logs worker1
```

**Restart services:**
```bash
docker-compose restart
```

### Port conflicts

If ports are already in use, modify `docker-compose.yml`:

```yaml
ports:
  - "8081:8080"  # Change 8080 to 8081
```

### Database issues

Reset the database:
```bash
docker-compose down -v
docker-compose up -d postgres
# Wait 10 seconds
docker-compose up -d
```

### Worker not connecting

**Check network:**
```bash
docker network ls
docker network inspect distributeai-network
```

**Check coordinator is accessible:**
```bash
docker exec distributeai-worker1 curl http://coordinator:8080/health
```

---

## Production Deployment Considerations

For production use, consider:

1. **Security**:
   - Add TLS/SSL certificates
   - Implement authentication (JWT)
   - Use secrets management (Vault)
   - Restrict network access

2. **High Availability**:
   - Deploy multiple coordinators behind load balancer
   - Use managed PostgreSQL (RDS, Cloud SQL)
   - Redis Cluster for queue
   - Multi-region worker deployment

3. **Monitoring**:
   - Configure alerting in Grafana
   - Set up log aggregation (ELK stack)
   - Add health check endpoints

4. **Scaling**:
   - Kubernetes deployment (see `deployments/k8s/`)
   - Auto-scaling based on job queue depth
   - Regional worker pools

---

## Development Mode

For local development:

```bash
# Start only infrastructure (DB, Redis, MinIO)
docker-compose up -d postgres redis minio

# Run coordinator locally
cd coordinator
go run ./cmd/coordinator

# Run worker locally
cd worker
COORDINATOR_URL=http://localhost:8080 go run ./cmd/worker

# Run dashboard locally
cd dashboard
npm install
npm run dev
```

---

## Kubernetes Deployment (Advanced)

For Kubernetes deployment:

```bash
# Apply manifests
kubectl apply -f deployments/k8s/

# Check status
kubectl get pods -n distributeai

# Access services
kubectl port-forward svc/coordinator 8080:8080
kubectl port-forward svc/dashboard 3000:3000
```

---

## Support

- **Documentation**: See `docs/` directory
- **Issues**: GitHub Issues
- **Architecture**: `docs/ARCHITECTURE.md`
- **Demo Guide**: `docs/DEMO_GUIDE.md`

---

## Quick Reference

```bash
# Start platform
./run.sh

# View logs
docker-compose logs -f coordinator

# Submit job
curl -X POST http://localhost:8080/api/v1/jobs -H "Content-Type: application/json" -d @examples/hash-verify/job.json

# View dashboard
open http://localhost:3000

# Stop platform
docker-compose down
```
