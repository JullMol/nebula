# 🌌 Nebula

<div align="center">

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![gRPC](https://img.shields.io/badge/gRPC-244c5a?style=for-the-badge&logo=google&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-DC382D?style=for-the-badge&logo=redis&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)

**A lightweight, distributed serverless code execution platform**

*Execute Python & Node.js code in isolated Docker containers with automatic load balancing*

[Features](#-features) • [Architecture](#-architecture) • [Getting Started](#-getting-started) • [API Reference](#-api-reference)

</div>

---

## ⚡ Features

| Feature | Description |
|---------|-------------|
| 🐳 **Container Isolation** | Each code execution runs in a fresh Docker container |
| ⚖️ **Load Balancing** | Round-robin distribution across multiple worker nodes |
| 🔄 **Async Job Queue** | Redis-powered job queue with background processing |
| 📊 **Real-time Monitoring** | Prometheus metrics + Grafana dashboards |
| 🌐 **gRPC Communication** | High-performance inter-service communication |
| 🐍 **Multi-Language** | Support for Python and Node.js |

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                           CLIENT (Browser)                          │
└─────────────────────────────────┬───────────────────────────────────┘
                                  │ HTTP/REST
                                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         API GATEWAY (Fiber)                         │
│  • Authentication (API Key)      • Rate Limiting                    │
│  • Job Submission                • Status Polling                   │
│  • Prometheus Metrics            • Static UI Serving                │
└───────────┬─────────────────────────────────────────┬───────────────┘
            │                                         │
            ▼                                         ▼
┌───────────────────────┐                 ┌───────────────────────────┐
│      PostgreSQL       │                 │          Redis            │
│   (Job Persistence)   │                 │      (Job Queue)          │
└───────────────────────┘                 └─────────────┬─────────────┘
                                                        │
                          ┌─────────────────────────────┼─────────────┐
                          │                             │             │
                          ▼                             ▼             ▼
              ┌───────────────────┐         ┌───────────────────┐
              │    WORKER 1       │         │    WORKER 2       │  ...
              │  (gRPC Server)    │         │  (gRPC Server)    │
              │                   │         │                   │
              │  ┌─────────────┐  │         │  ┌─────────────┐  │
              │  │   Docker    │  │         │  │   Docker    │  │
              │  │  Container  │  │         │  │  Container  │  │
              │  └─────────────┘  │         │  └─────────────┘  │
              └───────────────────┘         └───────────────────┘
```

---

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- Docker Desktop
- PostgreSQL (via Docker)
- Redis (via Docker)

### Quick Start

```bash
# Clone the repository
git clone https://github.com/JullMol/nebula.git
cd nebula

# Start infrastructure
docker-compose up -d redis postgres prometheus grafana

# Run Gateway (Terminal 1)
go run cmd/gateway/main.go

# Run Workers (Terminal 2 & 3)
go run cmd/worker/main.go -port 9091
go run cmd/worker/main.go -port 9092
```

Open **http://localhost:3000** in your browser 🎉

---

## 📁 Project Structure

```
nebula/
├── api/
│   ├── proto/              # Protocol Buffer definitions
│   └── pb/                 # Generated gRPC code
├── cmd/
│   ├── gateway/            # API Gateway service
│   │   ├── main.go
│   │   └── index.html      # Web UI
│   └── worker/             # Worker service
│       └── main.go
├── internal/
│   ├── gateway/
│   │   └── proxy/          # gRPC proxy to workers
│   ├── orchestrator/
│   │   └── scheduler/      # Load balancer (Round Robin)
│   ├── platform/
│   │   ├── database/       # PostgreSQL connection
│   │   ├── docker/         # Docker client
│   │   └── queue/          # Redis queue
│   └── worker/             # Worker gRPC server
├── pkg/
│   └── config/             # Configuration loader
├── deploy/
│   ├── docker/             # Dockerfiles
│   └── monitoring/         # Prometheus config
├── docker-compose.yaml
└── config.yaml
```

---

## 📡 API Reference

### Submit Job

```bash
POST /submit
Headers: X-API-KEY: rahasia-negara
Content-Type: application/json

{
  "image": "python:alpine",
  "command": "",
  "code": "print('Hello, Nebula!')"
}
```

**Response:**
```json
{
  "status": "queued",
  "job_id": "uuid-here",
  "info": "Job tersimpan di DB & Masuk Redis"
}
```

### Check Status

```bash
GET /status/:job_id
```

**Response:**
```json
{
  "job_id": "uuid-here",
  "status": "completed",
  "result": "Hello, Nebula!\n",
  "created_at": "2024-01-05T10:00:00Z",
  "updated_at": "2024-01-05T10:00:03Z"
}
```

---

## 📊 Monitoring

Access dashboards:
- **Prometheus:** http://localhost:9090
- **Grafana:** http://localhost:4000 (admin/admin)

Available metrics:
- `nebula_jobs_submitted_total` - Total jobs submitted
- `nebula_jobs_processed_total{status="completed|failed"}` - Jobs by status

---

## 🛠️ Tech Stack

| Layer | Technology |
|-------|------------|
| Language | Go 1.21 |
| Web Framework | Fiber v2 |
| RPC | gRPC + Protocol Buffers |
| Database | PostgreSQL 15 |
| Queue | Redis |
| Container | Docker |
| Monitoring | Prometheus + Grafana |

---

## 🎯 How It Works

1. **Client** submits code via REST API or Web UI
2. **Gateway** saves job to PostgreSQL and pushes to Redis queue
3. **Background Dispatcher** dequeues job and forwards to available worker
4. **Worker** creates temp file, mounts to Docker container, executes
5. **Worker** captures logs and returns result via gRPC
6. **Gateway** updates PostgreSQL with result
7. **Client** polls status endpoint until completion

---

## 📝 License

MIT License - feel free to use for learning and portfolio!

---

<div align="center">
  
**Built with ❤️ by [JullMol](https://github.com/JullMol)**

</div>
