# 🏗️ Architecture Documentation - DistributeAI

## System Overview

DistributeAI is a decentralized compute platform consisting of multiple microservices that work together to schedule, execute, and verify computational jobs across a distributed network of worker nodes.

---

## Core Components

### 1. Coordinator (Control Plane)

**Technology**: Go 1.21
**Port**: 8080
**Database**: PostgreSQL
**Queue**: Redis
**Storage**: MinIO (S3-compatible)

#### Responsibilities:
- Accept job submissions via REST API
- Maintain registry of available worker nodes
- Schedule jobs to appropriate workers
- Perform k-of-n verification on results
- Track node reputation and statistics
- Handle fault tolerance and rescheduling

#### Internal Modules:

**`internal/models/`** - Data structures
- `Job`: Represents a computational task
- `Node`: Represents a worker machine
- `JobExecution`: Tracks job execution on a specific node
- `VerificationResult`: Results of k-of-n consensus

**`internal/repository/`** - Database layer
- PostgreSQL schema management
- CRUD operations for jobs, nodes, executions
- Transaction handling

**`internal/scheduler/`** - Job scheduling engine
- Polls for pending jobs every 5 seconds
- Selects workers based on:
  - Resource availability (CPU, memory, GPU)
  - Reputation score (higher is better)
  - Current workload
- Creates job executions for redundancy
- Monitors running jobs
- Detects stale nodes (no heartbeat for 2+ minutes)

**`internal/verification/`** - k-of-n verification engine
- Collects results from multiple executions
- Compares result hashes
- Determines consensus (e.g., 2 out of 3 must agree)
- Updates node reputations:
  - +5 for agreeing with consensus
  - -10 for disagreeing
- Finalizes job status

**`internal/api/`** - REST API handlers
- Job submission and retrieval
- Node registration and heartbeats
- Result submission from workers
- System statistics

#### API Endpoints:

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/stats` | System statistics |
| `POST` | `/api/v1/jobs` | Submit new job |
| `GET` | `/api/v1/jobs` | List all jobs |
| `GET` | `/api/v1/jobs/:id` | Get job details |
| `GET` | `/api/v1/jobs/:id/executions` | Get job executions |
| `POST` | `/api/v1/nodes/register` | Register worker node |
| `GET` | `/api/v1/nodes` | List all nodes |
| `GET` | `/api/v1/nodes/:id` | Get node details |
| `POST` | `/api/v1/nodes/:id/heartbeat` | Worker heartbeat |
| `GET` | `/api/v1/nodes/:nodeId/pending-jobs` | Get pending jobs for worker |
| `POST` | `/api/v1/worker/result` | Submit job result |
| `GET` | `/metrics` | Prometheus metrics |

---

### 2. Worker Agent

**Technology**: Go 1.21
**Execution**: Docker containers
**Dependencies**: Docker Engine

#### Responsibilities:
- Register with coordinator on startup
- Send heartbeats every 30 seconds
- Poll for pending jobs every 10 seconds
- Execute jobs in isolated Docker containers
- Compute SHA256 hash of results
- Submit results back to coordinator

#### Internal Modules:

**`internal/client/`** - Coordinator API client
- HTTP client for coordinator communication
- Request/response models
- Retry logic for network failures

**`internal/executor/`** - Docker job executor
- Docker SDK integration
- Container lifecycle management:
  1. Pull image
  2. Create container with resource limits
  3. Start container
  4. Wait for completion (5-minute timeout)
  5. Collect logs
  6. Compute result hash
  7. Cleanup
- Resource constraints:
  - Memory: 512MB default
  - CPU: 1 core default
  - Auto-remove containers after execution

**`internal/monitor/`** - System monitoring
- CPU usage tracking
- Memory usage tracking
- Hostname detection

#### Worker Lifecycle:

```
Start → Register → [Heartbeat Loop] → [Job Polling Loop] → Shutdown
                          ↓                    ↓
                    Every 30s            Every 10s
                          ↓                    ↓
                   Update Status      Execute Jobs → Submit Results
```

---

### 3. CLI Tool

**Technology**: Go 1.21 + Cobra
**Binary**: `distributeai`

#### Commands:

| Command | Description | Example |
|---------|-------------|---------|
| `submit` | Submit new job | `distributeai submit --name "Test" --image "alpine:latest" --cmd "echo" --cmd "hello"` |
| `list` | List all jobs | `distributeai list` |
| `get <id>` | Get job details | `distributeai get abc-123` |
| `nodes` | List worker nodes | `distributeai nodes` |
| `stats` | System statistics | `distributeai stats` |

#### Features:
- Table-formatted output
- Color-coded status
- Auto-truncation of long IDs
- JSON parsing and pretty printing

---

### 4. Dashboard (Web UI)

**Technology**: React 18 + Vite
**Port**: 3000
**Deployment**: Nginx

#### Features:
- Real-time updates (5-second polling)
- System statistics cards:
  - Total nodes (online/busy)
  - Total CPU/memory resources
  - Jobs completed/running/failed
- Worker nodes list:
  - Status badges (online/busy/offline)
  - Reputation scores
  - Success rates
  - Resource details
- Recent jobs list:
  - Status tracking
  - Execution details
  - Result preview

#### Components:

**`App.jsx`** - Main application
- Data fetching from coordinator
- Auto-refresh every 5 seconds
- Error handling

**`StatsCards.jsx`** - Statistics display
- Node count and status
- Resource totals
- Job metrics

**`NodesList.jsx`** - Worker nodes grid
- Node cards with status badges
- Reputation and statistics
- Resource information

**`JobsList.jsx`** - Recent jobs feed
- Job cards with status
- Docker image and config
- Result preview

---

## Data Flow

### Job Submission Flow

```
1. User submits job
   ↓
2. Coordinator creates Job record (status: pending)
   ↓
3. Scheduler selects 3 workers (based on reputation & resources)
   ↓
4. Coordinator creates 3 JobExecution records (status: scheduled)
   ↓
5. Workers poll and receive pending jobs
   ↓
6. Workers execute jobs in Docker containers
   ↓
7. Workers compute SHA256(result) and submit back
   ↓
8. Coordinator collects results
   ↓
9. Verification engine compares hashes
   ↓
10. If 2/3 match → Consensus reached
    ├─ Job marked as completed
    ├─ Agreeing nodes: +5 reputation
    └─ Disagreeing nodes: -10 reputation
```

### Heartbeat Flow

```
Worker → POST /nodes/:id/heartbeat (every 30s)
  ├─ Includes: CPU usage, memory usage, active jobs
  └─ Coordinator updates last_heartbeat timestamp

Coordinator Scheduler (every 5s)
  ├─ Checks all nodes
  └─ If last_heartbeat > 2 minutes ago:
      ├─ Mark node as offline
      └─ Penalize reputation (-20)
```

---

## Database Schema

### Jobs Table
```sql
CREATE TABLE jobs (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    docker_image VARCHAR(255) NOT NULL,
    command JSONB NOT NULL,
    environment JSONB,
    input_data TEXT,
    required_cpu INTEGER DEFAULT 1,
    required_memory INTEGER DEFAULT 1,
    required_gpu BOOLEAN DEFAULT FALSE,
    redundancy INTEGER DEFAULT 3,
    consensus INTEGER DEFAULT 2,
    status VARCHAR(50) NOT NULL,
    submitted_by VARCHAR(255),
    submitted_at TIMESTAMP NOT NULL,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    result TEXT,
    error_message TEXT,
    credits_required INTEGER DEFAULT 1
);
```

### Nodes Table
```sql
CREATE TABLE nodes (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    region VARCHAR(100),
    cpu_cores INTEGER NOT NULL,
    memory_gb INTEGER NOT NULL,
    gpu_enabled BOOLEAN DEFAULT FALSE,
    gpu_model VARCHAR(255),
    status VARCHAR(50) NOT NULL,
    reputation_score REAL DEFAULT 100.0,
    total_jobs_run INTEGER DEFAULT 0,
    successful_jobs_run INTEGER DEFAULT 0,
    failed_jobs INTEGER DEFAULT 0,
    credits_earned INTEGER DEFAULT 0,
    last_heartbeat TIMESTAMP NOT NULL,
    registered_at TIMESTAMP NOT NULL,
    current_job_id VARCHAR(36)
);
```

### Job Executions Table
```sql
CREATE TABLE job_executions (
    id VARCHAR(36) PRIMARY KEY,
    job_id VARCHAR(36) NOT NULL REFERENCES jobs(id),
    node_id VARCHAR(36) NOT NULL REFERENCES nodes(id),
    status VARCHAR(50) NOT NULL,
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    result TEXT,
    result_hash VARCHAR(64),
    error_message TEXT,
    logs TEXT
);
```

---

## k-of-n Verification Algorithm

### Configuration
- **n (redundancy)**: Number of nodes to execute the job (default: 3)
- **k (consensus)**: Number of nodes that must agree (default: 2)

### Process

1. **Execution Phase**
   - Job scheduled to `n` workers
   - Each worker executes independently
   - No communication between workers

2. **Result Collection**
   - Workers submit: `(result, SHA256(result), logs)`
   - Coordinator waits for at least `k` completions

3. **Hash Comparison**
   ```go
   resultCounts := map[string]int{}
   for exec in completedExecutions:
       resultCounts[exec.ResultHash]++

   consensusHash := findMostCommon(resultCounts)
   if resultCounts[consensusHash] >= k:
       consensus = true
   ```

4. **Reputation Update**
   ```go
   for node in agreementNodes:
       reputation += 5

   for node in disagreementNodes:
       reputation -= 10
   ```

### Example Scenario

**Job**: Compute `echo "hello" | sha256sum`

**Execution**:
- Node A: `5891b5b522d5...` ✅
- Node B: `5891b5b522d5...` ✅
- Node C: `deadbeef1234...` ❌

**Verification**:
- Hash `5891b5b522d5...`: 2 votes
- Hash `deadbeef1234...`: 1 vote
- **Consensus reached** (2 ≥ 2)
- Result: `5891b5b522d5...`

**Reputation Changes**:
- Node A: +5
- Node B: +5
- Node C: -10

---

## Fault Tolerance

### Worker Failure Scenarios

| Scenario | Detection | Response |
|----------|-----------|----------|
| Worker crashes during job | No result submitted within timeout | Scheduler sees incomplete executions, consensus may still be reached with remaining nodes |
| Worker stops sending heartbeats | Heartbeat missed for 2+ minutes | Mark offline, penalize reputation (-20) |
| Worker submits wrong result | Result hash doesn't match consensus | Penalize reputation (-10), exclude from result |
| Worker becomes slow | Job timeout (5 minutes) | Mark execution as failed, use other nodes |

### Coordinator Failure
- **Current**: Single point of failure
- **Production**: Deploy multiple coordinators behind load balancer
- **State**: PostgreSQL provides persistence
- **Queue**: Redis can be clustered

---

## Security Considerations

### Current Implementation
- ✅ Docker container isolation
- ✅ Resource limits (CPU, memory)
- ✅ Result verification via hashing
- ✅ Auto-remove containers after execution

### Production Enhancements
- 🔒 Job signing (verify submitter identity)
- 🔒 Encrypted inputs (sensitive data)
- 🔒 TEE support (trusted execution environments)
- 🔒 Network isolation (prevent internet access in containers)
- 🔒 Node authentication (TLS certificates)
- 🔒 Rate limiting (prevent DoS)

---

## Scalability

### Horizontal Scaling

| Component | Scaling Strategy | Notes |
|-----------|------------------|-------|
| Coordinator | Load balancer + multiple instances | Requires shared PostgreSQL/Redis |
| Workers | Add more nodes dynamically | Linear scaling |
| Database | PostgreSQL replication | Read replicas for queries |
| Queue | Redis Cluster | Partition job queue |
| Dashboard | CDN + multiple instances | Stateless |

### Performance Targets

| Metric | Current | Target (Production) |
|--------|---------|---------------------|
| Jobs/hour | 100 | 10,000+ |
| Worker nodes | 3 | 1,000+ |
| Job latency (simple) | 30s | 10s |
| Consensus verification | 15s | 5s |
| Coordinator response time | < 100ms | < 50ms |

---

## Monitoring & Observability

### Prometheus Metrics

The coordinator exposes metrics at `/metrics`:

- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request latency
- `jobs_total` - Total jobs by status
- `nodes_total` - Total nodes by status
- `job_execution_duration_seconds` - Job execution time
- `verification_consensus_reached` - Consensus success rate

### Grafana Dashboards

Pre-configured dashboards for:
- System overview (nodes, jobs, resources)
- Job performance (latency, success rate)
- Node health (uptime, reputation trends)
- Resource utilization (CPU, memory)

### Logging

- Coordinator: JSON logs to stdout
- Workers: JSON logs to stdout
- Docker containers: Captured and stored in job_executions table

---

## Deployment

### Development
```bash
./run.sh
```

### Production

**Kubernetes Deployment** (future):
```yaml
- Coordinator: Deployment with 3 replicas
- Workers: DaemonSet or separate deployments
- PostgreSQL: StatefulSet with persistent volumes
- Redis: StatefulSet with persistence
- Dashboard: Deployment with CDN
```

**Cloud Providers**:
- AWS: ECS/EKS
- Azure: AKS
- GCP: GKE
- Self-hosted: Docker Swarm or Kubernetes

---

## Future Enhancements

1. **WebSocket Support**: Real-time dashboard updates
2. **Payment Integration**: Stripe/crypto for job payments
3. **Smart Contracts**: On-chain verification and payments
4. **P2P Discovery**: LibP2P for node discovery
5. **ZK Proofs**: Zero-knowledge result verification
6. **GPU Support**: NVIDIA/AMD GPU scheduling
7. **Spot Instances**: Cost optimization
8. **Auto-scaling**: Dynamic worker provisioning
9. **ML Model Registry**: Store and distribute models
10. **Multi-tenancy**: Isolated networks per user

---

## Development

### Running Tests
```bash
cd coordinator
go test ./...

cd ../worker
go test ./...

cd ../cli
go test ./...
```

### Building
```bash
# Build all components
make build

# Or individually
cd coordinator && go build ./cmd/coordinator
cd worker && go build ./cmd/worker
cd cli && go build ./cmd/distributeai
```

### Debugging
```bash
# Coordinator logs
docker-compose logs -f coordinator

# Worker logs
docker-compose logs -f worker1

# Database query
docker-compose exec postgres psql -U distributeai
```

---

## References

- [Docker SDK for Go](https://docs.docker.com/engine/api/sdk/)
- [Gin Web Framework](https://gin-gonic.com/)
- [PostgreSQL](https://www.postgresql.org/)
- [Redis](https://redis.io/)
- [Prometheus](https://prometheus.io/)
- [React](https://react.dev/)
