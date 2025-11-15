# Image Processing Demo

This demonstrates a semi-deterministic ML workload using Python.

## Purpose

Shows how AI/ML workloads can be distributed across the network with verification.

## Expected Result

All nodes should produce identical hash outputs when using the same seed.

## How to Run

```bash
# Using the CLI
./cli/bin/distributeai submit \
  --name "Image Processing Test" \
  --image "python:3.9-slim" \
  --cpu 2 \
  --memory 2 \
  --cmd "python" \
  --cmd "-c" \
  --cmd "import hashlib; data = 'DistributeAI-Image-Process'; print(hashlib.sha256(data.encode()).hexdigest())"

# Or submit the pre-defined job
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d @examples/image-process/job.json
```

## Scaling

This example can be extended to:
- Process image datasets in parallel
- Run AI inference across multiple nodes
- Perform distributed model training
