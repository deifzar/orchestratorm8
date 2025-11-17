# OrchestratorM8 - Security Services Orchestration Platform

<div align="center">

**Production-grade Go microservice for coordinating distributed security scanning operations.**

[![Go Version](https://img.shields.io/badge/Go-1.21.5+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Enabled-2496ED?logo=docker)](dockerfile)
[![Status](https://img.shields.io/badge/status-active--development-yellow)](https://github.com/yourusername/orchestratorm8)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

[Features](#key-features) • [Quick Start](#quick-start) • [Documentation](#documentation) • [Architecture](#architecture) • [API](#api-reference)

</div>

---

## Overview

OrchestratorM8 is a production-grade Go microservice that acts as the central coordinator for the CPT (Continuous Penetration Testing) platform. It orchestrates complex security scanning workflows across multiple specialized microservices through RabbitMQ message queuing, managing scan workflows and coordinating between various security analysis tools.

**Built for:**
- Security operations teams managing continuous scanning workflows
- DevSecOps teams integrating security into CI/CD pipelines
- Penetration testing teams requiring coordinated multi-tool assessments
- Security researchers conducting large-scale reconnaissance operations

### Key Features

- **Centralized Orchestration**: Coordinates message flows between 5+ security scanning microservices
- **Advanced Connection Pooling**: RabbitMQ connection management with health monitoring and auto-reconnection
- **Database-Driven Workflows**: PostgreSQL-backed domain targeting with 15-minute polling intervals
- **Production Ready**: Docker containerization, health checks, graceful shutdown, and security hardening
- **Event-Driven Architecture**: Topic-based message routing with manual acknowledgment for reliability
- **Comprehensive Logging**: Zerolog with log rotation, Elasticsearch integration, and structured output

---

## Quick Start

### Prerequisites

- **Go** 1.21.5 or higher
- **PostgreSQL** 12+ (for domain management)
- **RabbitMQ** 3.8+ (for message queuing)
- **Docker** (optional, for containerized deployment)

### Installation

#### Option 1: Build from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/OrchestratorM8.git
cd OrchestratorM8

# Install Go dependencies
go mod download

# Build the binary
go build -o bin/orchestratorm8 .

# Run the service
./bin/orchestratorm8 launch --ip-address 0.0.0.0 --port 8080
```

#### Option 2: Docker

```bash
# Build the Docker image
docker build -t orchestratorm8:latest .

# Run the container
docker run -d \
  -p 8080:8080 \
  -e POSTGRESQL_HOSTNAME=your-db-host \
  -e POSTGRESQL_DB=cpt_orchestrator \
  -e POSTGRESQL_USERNAME=cpt_dbuser \
  -e POSTGRESQL_PASSWORD=your-db-password \
  -e RABBITMQ_HOSTNAME=your-rabbitmq-host \
  -e RABBITMQ_USERNAME=orchestrator \
  -e RABBITMQ_PASSWORD=your-rabbitmq-password \
  --name orchestratorm8 \
  orchestratorm8:latest launch
```

### Configuration

1. Copy the example configuration:
```bash
cp configs/configuration_template.yaml configs/configuration.yaml
```

2. Edit `configs/configuration.yaml` with your settings:
```yaml
APP_ENV: PROD
LOG_LEVEL: "1"  # 0=debug, 1=info, 2=warn, 3=error

ORCHESTRATORM8:
  Services:
    asmm8: "${ASMM8_URL}"
    naabum8: "${NAABUM8_URL}"
    katanam8: "${KATANAM8_URL}"
    num8: "${NUM8_URL}"
  Exchanges:
    cptm8: "topic"
    scheduler: "topic"
    notification: "topic"

Database:
  location: "${POSTGRESQL_HOSTNAME}"
  port: 5432
  schema: "public"
  database: "${POSTGRESQL_DB}"
  username: "${POSTGRESQL_USERNAME}"
  password: "${POSTGRESQL_PASSWORD}"

RabbitMQ:
  location: "${RABBITMQ_HOSTNAME}"
  port: 5672
  username: "${RABBITMQ_USERNAME}"
  password: "${RABBITMQ_PASSWORD}"
  pool:
    max_connections: 10      # Maximum pool size
    min_connections: 2       # Minimum pool size
    health_check_period: "30m"
    connection_timeout: "30s"
    retry_attempts: 3
    retry_delay: "2s"
```

3. Set environment variables for sensitive data:
```bash
export POSTGRESQL_HOSTNAME="localhost"
export POSTGRESQL_DB="cpt_orchestrator"
export POSTGRESQL_USERNAME="cpt_dbuser"
export POSTGRESQL_PASSWORD="your-secure-password"

export RABBITMQ_HOSTNAME="localhost"
export RABBITMQ_USERNAME="orchestrator"
export RABBITMQ_PASSWORD="your-rabbitmq-password"

export ASMM8_URL="http://localhost:8000"
export NAABUM8_URL="http://localhost:8001"
export KATANAM8_URL="http://localhost:8002"
export NUM8_URL="http://localhost:8003"
```

### Database Setup

```sql
-- Create database
CREATE DATABASE cpt_orchestrator;

-- Create user
CREATE USER cpt_dbuser WITH PASSWORD 'your-secure-password';
GRANT ALL PRIVILEGES ON DATABASE cpt_orchestrator TO cpt_dbuser;

-- Connect to the database
\c cpt_orchestrator

-- Create tables
CREATE TABLE cptm8domain (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain VARCHAR(255) NOT NULL UNIQUE,
    company VARCHAR(255) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_cptm8domain_enabled ON cptm8domain(enabled);
CREATE INDEX idx_cptm8domain_company ON cptm8domain(company);

-- Sample data
INSERT INTO cptm8domain (domain, company, enabled) VALUES
('example.com', 'Example Corp', true),
('test-site.org', 'Test Organization', true);
```

---

## Usage

### Basic Workflow

1. **Start the orchestrator:**
```bash
./bin/orchestratorm8 launch --ip-address 0.0.0.0 --port 8080
```

2. **Check health status:**
```bash
# Liveness check
curl http://localhost:8080/health

# Readiness check (verifies DB and RabbitMQ)
curl http://localhost:8080/ready
```

3. **Add domains to the database:**
```sql
INSERT INTO cptm8domain (domain, company, enabled)
VALUES ('target.com', 'Target Company', true);
```

4. **The orchestrator will automatically:**
   - Poll the database every 15 minutes
   - Identify enabled domains
   - Publish scan initiation messages to RabbitMQ
   - Coordinate security services (asmm8, naabum8, katanam8, num8)

### Orchestration Modes

The orchestrator runs in a **continuous mode** with two concurrent goroutines:

#### Database Polling (15-minute intervals)
- Queries `cptm8domain` table for enabled domains
- Sends domain lists to the publish goroutine via channel
- Handles graceful shutdown on SIGINT/SIGTERM

#### Message Publishing
- Receives domain lists from polling goroutine
- Publishes scan initiation messages to RabbitMQ exchanges
- Coordinates workflows across security services

---

## API Reference

### Health Check Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Liveness probe (always 200 OK) |
| `GET` | `/ready` | Readiness probe (checks DB + RabbitMQ) |

**Example Response (`/ready`):**
```json
{
  "status": "ready",
  "database": "connected",
  "rabbitmq": "connected",
  "timestamp": "2025-11-17T12:34:56Z"
}
```

---

## Architecture

### High-Level Overview

```
┌─────────────┐
│   Client    │
│  (Admin)    │
└─────────────┘
       │
       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   PostgreSQL    │    │   OrchestratorM8 │    │    RabbitMQ     │
│   Database      │◄──►│   (Port 8080)    │◄──►│  Message Queue  │
│ (Domain Mgmt)   │    │  Gin/Go Service  │    │  (AMQP 0.9.1)   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                │ Publishes messages to queues
                                ▼
                   ┌─────────────────────────────┐
                   │     Security Services       │
                   │                             │
                   │ ┌─────────┐ ┌─────────────┐ │
                   │ │ asmm8   │ │   naabum8   │ │
                   │ │:8000    │ │   :8001     │ │
                   │ │Asset Mgmt│ │Network Scan│ │
                   │ └─────────┘ └─────────────┘ │
                   │                             │
                   │ ┌─────────┐ ┌─────────────┐ │
                   │ │katanam8 │ │    num8     │ │
                   │ │:8002    │ │   :8003     │ │
                   │ │Vuln Scan│ │  Analysis   │ │
                   │ └─────────┘ └─────────────┘ │
                   │                             │
                   │    ┌─────────────────┐      │
                   │    │  reportingm8    │      │
                   │    │  Report Gen     │      │
                   │    └─────────────────┘      │
                   └─────────────────────────────┘
```

### Orchestration Workflow

**Complete Orchestration Pipeline:**

```
1. DATABASE POLLING (Every 15 minutes)
   ├─→ Query enabled domains from PostgreSQL
   ├─→ Send domain list to publish goroutine
   └─→ Cleanup temporary files (24+ hours old)

2. MESSAGE PUBLISHING
   ├─→ Receive domain list from polling goroutine
   ├─→ Create scan initiation messages
   ├─→ Publish to RabbitMQ exchanges (cptm8, scheduler, notification)
   └─→ Route messages to appropriate service queues

3. SERVICE COORDINATION
   ├─→ asmm8 Queue: Asset discovery and subdomain enumeration
   ├─→ naabum8 Queue: Network analysis and port scanning
   ├─→ katanam8 Queue: Vulnerability assessment
   ├─→ num8 Queue: Numerical analysis and statistics
   └─→ reportingm8 Queue: Report generation and aggregation

4. HEALTH MONITORING
   ├─→ RabbitMQ connection health checks (30-minute intervals)
   ├─→ Consumer health tracking
   └─→ Automatic reconnection on failures
```

### Package Structure

```
orchestratorm8/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Base command setup
│   └── launch.go          # Service launcher command
├── pkg/                    # 8 packages, 22 Go files
│   ├── amqpM8/            # RabbitMQ connection pooling (5 files, 1,777 lines)
│   │   ├── pooled_amqp.go        # Main pooled AMQP implementation (842 lines)
│   │   ├── connection_pool.go    # Connection pool management (406 lines)
│   │   ├── pool_manager.go       # Global pool manager (190 lines)
│   │   ├── shared_state.go       # Shared state management (241 lines)
│   │   └── initialization.go     # Pool initialization (98 lines)
│   ├── api8/              # HTTP API routes and initialization
│   ├── cleanup8/          # Temporary file cleanup utilities
│   ├── configparser/      # Configuration management (Viper)
│   ├── controller8/       # Orchestration controllers (267 lines)
│   ├── db8/               # Database access layer (PostgreSQL)
│   ├── log8/              # Structured logging (zerolog)
│   ├── model8/            # Data models and domain entities
│   └── utils/             # Utility functions
├── configs/               # Configuration files
│   ├── configuration_template.yaml  # Template with env vars
│   └── examples/          # Example configurations
├── docs/                  # Comprehensive documentation (6 files)
├── bin/                   # Compiled binaries
└── main.go                # Application entry point (sets umask 0027)
```

### Key Components

- **[Orchestration Controller](pkg/controller8/)** - Main coordination logic (267 lines)
- **[AMQP Connection Pool](pkg/amqpM8/)** - Advanced RabbitMQ pooling (1,777 lines)
- **[Database Layer](pkg/db8/)** - PostgreSQL repository pattern
- **[API Layer](pkg/api8/)** - Gin-based HTTP endpoints
- **[Configuration Parser](pkg/configparser/)** - Viper-based config management
- **[Logging System](pkg/log8/)** - Zerolog with Lumberjack rotation

For detailed architecture documentation, see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

---

## Advanced Features

### RabbitMQ Integration

**Advanced Connection Pooling:**
- Configurable pool size (default: 2-10 connections)
- Automatic connection recovery on failures
- Periodic health checks (30-minute intervals)
- Manual message acknowledgment (autoack: false)
- Connection lifecycle management with context-aware shutdown

**Exchange Configuration:**
- **cptm8** (topic): Main security scanning coordination
- **scheduler** (topic): Scheduled operations and timing
- **notification** (topic): Event notifications and alerts

**Queue Configuration:**
- Prefetch count: 1 (ensures fair load balancing)
- Max length: 1 (prevents queue overflow)
- Overflow policy: "reject-publish" (prevents memory issues)
- Topic-based routing keys (e.g., `cptm8.asmm8.#`, `cptm8.naabum8.#`)

### Security Hardening

**Production Security Features:**
- No hardcoded credentials (environment variables only)
- File permission controls (umask 0027 = files: 0640, dirs: 0750)
- SQL injection protection via prepared statements
- Transaction rollback support
- Proper resource cleanup with deferred statements
- Log file permissions: 0640

### Error Handling

**Resilient Design:**
- Context-aware graceful shutdown (SIGINT/SIGTERM)
- Database connection retry logic
- RabbitMQ reconnection with exponential backoff
- Comprehensive error logging with stack traces
- Health check endpoints for monitoring

### Concurrent Processing

- Dual goroutine architecture (polling + publishing)
- Channel-based communication between goroutines
- Mutex-protected shared state in connection pool
- Non-blocking cleanup routines
- Thread-safe connection pool management

---

## Documentation

Comprehensive documentation is available in the [docs/](docs/) directory:

| Document | Description |
|----------|-------------|
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | Detailed system architecture and design patterns |
| [DEVELOPMENT.md](docs/DEVELOPMENT.md) | Development setup and guidelines |
| [SECURITY.md](docs/SECURITY.md) | Security best practices and hardening |
| [PERFORMANCE.md](docs/PERFORMANCE.md) | Performance optimization recommendations |
| [TODO.md](docs/TODO.md) | Production readiness roadmap and known issues |
| [CODE_REVIEW.md](docs/CODE_REVIEW.md) | Code quality analysis and recommendations |

---

## Security Services Coordinated

OrchestratorM8 coordinates the following security microservices:

| Service | Port | Purpose | Routing Key Pattern |
|---------|------|---------|-------------------|
| **asmm8** | 8000 | Asset discovery and subdomain enumeration | `cptm8.asmm8.#` |
| **naabum8** | 8001 | Network analysis and port scanning | `cptm8.naabum8.#` |
| **katanam8** | 8002 | Vulnerability scanning and assessment | `cptm8.katanam8.#` |
| **num8** | 8003 | Numerical analysis and statistics | `cptm8.num8.#` |
| **reportingm8** | - | Report generation and aggregation | `cptm8.reportingm8.#` |

All services communicate asynchronously through RabbitMQ message queues.

---

## Performance

**Typical Performance Metrics:**

- **Database Polling**: Every 15 minutes (configurable, but currently hardcoded)
- **RabbitMQ Connections**: Pool of 2-10 connections with health checks
- **Health Check Interval**: 30 minutes for consumer health monitoring
- **Connection Timeout**: 30 seconds with 3 retry attempts
- **Temporary File Cleanup**: Removes files older than 24 hours

**Resource Requirements:**

- **CPU**: 2+ cores recommended
- **Memory**: 1 GB minimum, 2 GB recommended
- **Storage**: 5 GB for application + logs + temporary files
- **Network**: Stable connectivity to PostgreSQL and RabbitMQ

For optimization tips, see [docs/PERFORMANCE.md](docs/PERFORMANCE.md).

---

## Security Considerations

### Current Security Posture

**Strengths:**
- All secrets via environment variables (no hardcoded credentials)
- File permissions enforced (umask 0027)
- SQL injection protection via prepared statements
- Transaction support with proper rollback
- Runs as non-root user in Docker containers

### Known Limitations

- No authentication on API endpoints (health checks are public)
- Database polling architecture (not event-driven)
- Polling interval is hardcoded (not configurable)
- No TLS/SSL enforcement for database or RabbitMQ connections
- Limited input validation

### Recommendations

1. **Enable TLS/SSL** for PostgreSQL and RabbitMQ connections
2. **Implement API authentication** (JWT or API keys)
3. **Add input validation** layer for all endpoints
4. **Configure event-driven architecture** (PostgreSQL LISTEN/NOTIFY)
5. **Implement rate limiting** for API endpoints
6. **Add security event audit logging**

See [docs/SECURITY.md](docs/SECURITY.md) for comprehensive security guidelines.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Follow Go best practices and existing code style
4. Add tests for new functionality (currently lacking - high priority!)
5. Update documentation as needed
6. Ensure no hardcoded credentials or secrets
7. Commit your changes (`git commit -m 'Add amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request to the `develop` branch

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for detailed development guidelines.

---

## Roadmap

### Version 1.x (Current)
- [x] Core orchestration functionality
- [x] PostgreSQL domain management
- [x] RabbitMQ message queuing with connection pooling
- [x] Docker containerization
- [x] Health check endpoints
- [x] Security hardening (file permissions, no hardcoded credentials)
- [x] Graceful shutdown handling

### Version 2.0 (Planned)
- [ ] Comprehensive test suite (target: 80% coverage)
- [ ] JWT-based API authentication
- [ ] Event-driven architecture (PostgreSQL LISTEN/NOTIFY)
- [ ] Configurable polling interval
- [ ] Database connection pooling configuration
- [ ] Prometheus metrics integration
- [ ] Kubernetes deployment manifests
- [ ] Horizontal scaling with leader election
- [ ] Configuration hot-reload capability
- [ ] TLS/SSL support for database and RabbitMQ

See [docs/TODO.md](docs/TODO.md) for the complete roadmap and known issues.

---

## Troubleshooting

### Common Issues

**1. Database connection failures**
```bash
# Check PostgreSQL is running
systemctl status postgresql

# Verify connection settings
psql -h localhost -U cpt_dbuser -d cpt_orchestrator

# Check PostgreSQL logs
sudo tail -f /var/log/postgresql/postgresql-*.log
```

**2. RabbitMQ connection errors**
```bash
# Check RabbitMQ status
systemctl status rabbitmq-server

# Verify credentials and connectivity
rabbitmqctl list_users
rabbitmqctl list_exchanges

# Access management interface
# http://localhost:15672 (default: guest/guest)
```

**3. No domains being processed**
```bash
# Check database has enabled domains
psql -U cpt_dbuser -d cpt_orchestrator -c "SELECT * FROM cptm8domain WHERE enabled = true;"

# Verify 15-minute polling interval has elapsed
# Check logs for detailed information
tail -f log/orchestratorm8.log
```

**4. Permission errors**
```bash
# Ensure log directory is writable
chmod 750 log/

# Check log file permissions
ls -la log/orchestratorm8.log  # Should be 0640

# Verify configuration file permissions
chmod 640 configs/configuration.yaml
```

---

## Technology Stack

| Component | Version | Purpose |
|-----------|---------|---------|
| **Go** | 1.21.5+ | Primary programming language |
| [**Gin**](https://github.com/gin-gonic/gin) | v1.10.1 | HTTP web framework |
| [**RabbitMQ AMQP**](https://github.com/rabbitmq/amqp091-go) | v1.10.0 | Message queue client |
| [**lib/pq**](https://github.com/lib/pq) | v1.10.9 | PostgreSQL driver |
| [**Zerolog**](https://github.com/rs/zerolog) | v1.33.0 | Structured logging |
| [**Lumberjack**](https://github.com/natefinch/lumberjack) | v2.2.1 | Log rotation |
| [**Cobra**](https://github.com/spf13/cobra) | v1.8.1 | CLI framework |
| [**Viper**](https://github.com/spf13/viper) | v1.19.0 | Configuration management |
| [**UUID**](https://github.com/gofrs/uuid) | v5.3.1 | UUID generation |

---

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [RabbitMQ](https://www.rabbitmq.com/) for reliable message queuing
- [PostgreSQL](https://www.postgresql.org/) for robust data persistence
- [Gin Web Framework](https://github.com/gin-gonic/gin) for the HTTP router
- The Go community for excellent open-source libraries

---

## Contact

For questions, issues, or feature requests, please open an issue on GitHub.

**Project Link:** [https://github.com/yourusername/OrchestratorM8](https://github.com/yourusername/OrchestratorM8)

---

<div align="center">

**Built with ❤️ for the security community**

[⬆ Back to Top](#orchestratorm8---security-services-orchestration-platform)

</div>
