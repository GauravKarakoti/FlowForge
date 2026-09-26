# 🌊 FlowForge

![Build Status](https://img.shields.io/badge/build-passing-brightgreen) ![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)

FlowForge is a high-performance, open-source distributed workflow execution platform. It empowers developers to create, schedule, queue, execute, and monitor asynchronous jobs across a fleet of worker nodes with enterprise-grade reliability. 

Whether you are processing massive data pipelines, sending millions of transactional emails, or orchestrating complex microservices, FlowForge guarantees that your code executes successfully, even in the face of infrastructure failures.

## ✨ Core Features

* **Distributed Execution:** Scale workloads horizontally across distributed worker nodes with zero single points of failure.
* **Advanced Scheduling:** Support for cron-based scheduling, delayed execution, interval triggers, and priority queuing.
* **Resilience & Fault Tolerance:** Built-in automatic retries, exponential backoff, timeout handling, and dead-letter queues (DLQ).
* **State Persistence:** Backed by robust persistent storage (PostgreSQL/Redis) ensuring no jobs are lost during crashes or restarts.
* **Granular Observability:** Real-time monitoring, job state tracking, and integration with Prometheus and OpenTelemetry.
* **Secure Authentication:** Role-Based Access Control (RBAC) and mTLS between the control plane and worker nodes.
* **Multi-Language SDKs:** Write workflows and workers in Go, Python, Node.js, or Java.

## 🚀 Quick Start

### 1. Start the FlowForge Control Plane
The easiest way to get started locally is using Docker Compose:

```bash
curl -O [https://raw.githubusercontent.com/GauravKarakoti/flowforge/main/docker-compose.yml](https://raw.githubusercontent.com/GauravKarakoti/flowforge/main/docker-compose.yml)
docker-compose up -d
```
*This spins up the FlowForge server, Redis queue, PostgreSQL state store, and the Web UI.*

### 2. Install the SDK (Python Example)
```bash
pip install flowforge-sdk
```

### 3. Write a Worker and a Job
```python
from flowforge import Worker, Client

# Define a task
def process_payment(amount, currency):
    print(f"Processing {amount} {currency}...")
    return True

# Start a worker node
worker = Worker(queue="payments")
worker.register_task("process_payment", process_payment)
worker.start()

# In another process, submit a job to the cluster
client = Client()
client.enqueue("process_payment", args=[150.00, "USD"], retry_limit=3)
```

## 📊 Dashboard & Observability
FlowForge comes with a built-in Web UI accessible at http://localhost:8080 by default. From the dashboard, you can:
  - Inspect queue depths and worker health.
  - View job logs, stack traces, and execution timelines.
  - Pause, resume, or manually trigger workflows.

## 🤝 Contributing
We welcome contributions of all sizes! Whether it's fixing a typo in the docs, reporting a bug, or building a new feature. Please see our [Contributing Guide](./CONTRIBUTING.md) to get started.