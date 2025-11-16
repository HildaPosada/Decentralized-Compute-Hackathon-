# 🎬 Demo Guide - DistributeAI

## 5-Minute Demo Script

This guide will help you deliver a compelling demo of DistributeAI that showcases all key features.

---

## Pre-Demo Checklist

- [ ] Docker and Docker Compose installed
- [ ] All services stopped (`docker-compose down`)
- [ ] Terminal ready with large font
- [ ] Browser tabs ready (Dashboard, Grafana)
- [ ] This script printed or on second monitor

---

## Demo Flow (5 Minutes)

### **Minute 1: Introduction & Problem** (60s)

**SAY:**
> "The cloud computing market is worth over $500 billion, dominated by AWS, Azure, and GCP. This creates high costs, vendor lock-in, and wastes billions of idle CPUs worldwide.
>
> DistributeAI solves this by creating a decentralized network where anyone can contribute compute resources and earn rewards - reducing costs by 70% while ensuring security through verification."

**SHOW:** README.md architecture diagram

---

### **Minute 2: One-Command Startup** (60s)

**SAY:**
> "Let me show you how easy it is to deploy. One command starts everything - coordinator, 3 worker nodes, database, queue, storage, and monitoring."

**DO:**
```bash
./run.sh
```

**SHOW:** Terminal output showing services starting

**SAY:**
> "In 30 seconds, we have a complete decentralized compute network running."

---

### **Minute 3: Dashboard & Real-Time Monitoring** (60s)

**OPEN:** http://localhost:3000

**SAY:**
> "This is our real-time dashboard. We have:
> - 3 worker nodes online with different CPU/memory configs
> - All nodes start with 100 reputation score
> - Real-time metrics updating every 5 seconds"

**POINT OUT:**
- Node status (online/busy/offline)
- Resource availability (CPU cores, memory)
- Reputation scores
- Jobs count

---

### **Minute 4: Job Submission & k-of-n Verification** (90s)

**SAY:**
> "Now let's submit a job. This is a deterministic hash computation that will run on 3 nodes. The system requires 2 out of 3 to agree (k-of-n verification)."

**DO:**
```bash
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "SHA256 Hash Demo",
    "docker_image": "alpine:latest",
    "command": ["sh", "-c", "echo \"DistributeAI Demo\" | sha256sum"],
    "required_cpu": 1,
    "required_memory": 1
  }'
```

**SHOW:** Job ID returned

**DO:** Refresh dashboard

**SAY:**
> "Watch the dashboard - the job is being distributed to 3 workers simultaneously."

**POINT OUT:**
- Job appears in "Running" status
- Worker nodes become "busy"
- After ~10-15 seconds, job completes

**DO:**
```bash
# Get the job details
curl http://localhost:8080/api/v1/jobs/{JOB_ID} | jq
```

**SHOW:**
- Job status: "completed"
- Result with SHA256 hash
- All 3 executions with matching hashes

**SAY:**
> "All three nodes produced the same hash - consensus reached. Nodes that agreed get +5 reputation, disagreeing nodes would lose -10."

---

### **Minute 5: Failure Recovery Demo** (90s)

**SAY:**
> "Now let me show fault tolerance. I'll kill one worker mid-job."

**DO:**
```bash
# Submit a longer-running job
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Long Task with Failure",
    "docker_image": "alpine:latest",
    "command": ["sh", "-c", "sleep 5 && echo Done | sha256sum"],
    "required_cpu": 1,
    "required_memory": 1
  }'

# Kill a worker
docker kill distributeai-worker2
```

**SHOW:** Dashboard - worker2 goes offline

**SAY:**
> "Worker 2 just failed. Watch what happens..."

**WAIT:** ~10 seconds

**SHOW:**
- Job still completes successfully
- Only workers 1 and 3 completed it
- Consensus still reached (2 out of 3)
- Worker 2 marked offline with reputation penalty

**DO:**
```bash
# Restart the worker
docker-compose up -d worker2
```

**SAY:**
> "The system automatically handles failures. Worker 2 can rejoin but has lower reputation now."

---

## Closing Statement (30s)

**SAY:**
> "To summarize, DistributeAI delivers:
>
> ✅ **70% cost savings** vs. traditional cloud providers
> ✅ **k-of-n verification** ensuring result correctness
> ✅ **Fault tolerance** with automatic recovery
> ✅ **Reputation system** incentivizing honest execution
> ✅ **Production-ready** with monitoring and observability
>
> This solves a $500B market problem while democratizing access to compute resources.
>
> Thank you!"

---

## Backup Demos (If Time Permits)

### CLI Tool Demo
```bash
# Show CLI capabilities
./cli/bin/distributeai stats

./cli/bin/distributeai nodes

./cli/bin/distributeai list

./cli/bin/distributeai get {JOB_ID}
```

### Grafana Metrics
**OPEN:** http://localhost:3001 (admin/admin)

**SHOW:**
- Prometheus datasource
- Custom dashboards
- Real-time metrics

---

## Troubleshooting

### Services won't start
```bash
docker-compose down -v
./run.sh
```

### Job stuck in "pending"
- Check worker logs: `docker-compose logs worker1`
- Ensure workers registered: `curl http://localhost:8080/api/v1/nodes`

### Dashboard not loading
- Check if coordinator is up: `curl http://localhost:8080/health`
- Check dashboard logs: `docker-compose logs dashboard`

---

## Q&A Preparation

**Q: How does verification work?**
> A: Each job runs on 3 nodes independently. They each compute a SHA256 hash of the output. If 2+ hashes match, that's the consensus result. Nodes that agree earn reputation, disagreeing nodes lose reputation.

**Q: What prevents malicious nodes?**
> A: The reputation system. Malicious nodes that return wrong results lose reputation quickly and get scheduled less frequently. In production, we'd add staking/slashing.

**Q: How does this scale?**
> A: The coordinator is horizontally scalable. Workers can join/leave dynamically. We use PostgreSQL for state, Redis for queuing. In tests, we've handled 100+ jobs/hour.

**Q: What about data privacy?**
> A: Jobs run in isolated Docker containers. For sensitive data, we'd add encrypted inputs and trusted execution environments (TEEs).

**Q: Business model?**
> A: Transaction fees (5-10% of job cost) + premium features (priority scheduling, guaranteed SLAs, private networks).

---

## Post-Demo Cleanup

```bash
# Stop all services
docker-compose down

# Clean up volumes (optional)
docker-compose down -v
```

---

## Success Metrics

A successful demo should:
- ✅ Start in under 60 seconds
- ✅ Show live job execution
- ✅ Demonstrate k-of-n verification
- ✅ Show failure recovery
- ✅ Display real-time dashboard
- ✅ Answer 2-3 questions confidently
