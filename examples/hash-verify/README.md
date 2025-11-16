# Hash Verification Demo

This is a deterministic workload that computes SHA256 hash of a text string.

## Purpose

Demonstrates k-of-n verification where multiple nodes execute the same job and must produce identical results.

## Expected Result

All nodes should produce the same SHA256 hash, demonstrating consensus.

## How to Run

```bash
# Using the CLI
./cli/bin/distributeai submit \
  --name "Hash Verification Test" \
  --image "alpine:latest" \
  --cmd "sh" \
  --cmd "-c" \
  --cmd "echo 'DistributeAI - Decentralized Compute for the People' | sha256sum"

# Or submit the pre-defined job
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d @examples/hash-verify/job.json
```

## Verification

The coordinator will:
1. Schedule the job to 3 worker nodes (redundancy=3)
2. Wait for results from all nodes
3. Compare result hashes
4. Reach consensus if 2+ nodes agree (consensus=2)
5. Reward agreeing nodes, penalize disagreeing nodes
