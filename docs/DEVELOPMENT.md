# OrchestratorM8 Development Guide

## Quick Start

### Prerequisites
- **Go**: Version 1.21.5 or later
- **PostgreSQL**: Version 12+ for database operations  
- **RabbitMQ**: Version 3.8+ for message queue operations
- **Docker**: For containerized deployment (optional)

### Initial Setup

1. **Clone and Setup**
   ```bash
   git clone <repository-url>
   cd OrchestratorM8
   go mod download
   ```

2. **Environment Configuration**
   ```bash
   # Copy template configuration
   cp configs/configuration_template.yaml configs/configuration.yaml

   # Set required environment variables
   export POSTGRESQL_HOSTNAME="localhost"
   export POSTGRESQL_DB="cpt_orchestrator"
   export POSTGRESQL_USERNAME="cpt_dbuser"
   export POSTGRESQL_PASSWORD="your_secure_password"

   export RABBITMQ_HOSTNAME="localhost"
   export RABBITMQ_USERNAME="orchestrator"
   export RABBITMQ_PASSWORD="your_rabbitmq_password"

   export ASMM8_URL="http://localhost:8000"
   export NAABUM8_URL="http://localhost:8001"
   export KATANAM8_URL="http://localhost:8002"
   export NUM8_URL="http://localhost:8003"
   ```

3. **Database Setup**
   ```sql
   -- Create database and user
   CREATE DATABASE cpt_orchestrator;
   CREATE USER cpt_dbuser WITH PASSWORD 'your_secure_password';
   GRANT ALL PRIVILEGES ON DATABASE cpt_orchestrator TO cpt_dbuser;
   
   -- Create required tables (see database schema section)
   ```

4. **Build and Run**
   ```bash
   # Build the application
   go build -o bin/orchestratorm8 .

   # Run with help
   ./bin/orchestratorm8 --help

   # Launch orchestrator (default: 0.0.0.0:8080)
   ./bin/orchestratorm8 launch

   # Launch with custom IP and port
   ./bin/orchestratorm8 launch --ip-address 127.0.0.1 --port 8090
   ```

## Project Structure

```
OrchestratorM8/
├── cmd/                    # CLI command definitions
│   ├── root.go            # Root Cobra command
│   └── launch.go          # Launch command for orchestrator
├── pkg/                    # Core application packages
│   ├── amqpM8/            # RabbitMQ connection pooling (5 files, 842 lines main)
│   ├── api8/              # API initialization and routing
│   ├── cleanup8/          # Temporary file cleanup utilities
│   ├── configparser/      # Configuration management
│   ├── controller8/       # Main orchestration controller
│   ├── db8/               # Database layer (PostgreSQL)
│   ├── log8/              # Logging with Zerolog + Lumberjack
│   ├── model8/            # Data models (Domain8, Notification8, etc.)
│   └── utils/             # Utility functions
├── configs/               # Configuration files directory
│   ├── configuration_template.yaml  # Template with env vars
│   └── examples/          # Example configurations
├── docs/                  # Documentation
├── bin/                   # Compiled binaries and Docker files
├── log/                   # Application logs (orchestratorm8.log)
├── tmp/                   # Temporary file storage
├── dockerfile             # Multi-stage Docker build
├── go.mod                 # Go module definition
└── main.go               # Application entry point (sets umask)
```

## Development Workflow

### Code Standards

#### Go Conventions
- **Package Naming**: Use lowercase, single-word package names
- **Interface Naming**: End interfaces with "-er" (e.g., `Publisher`, `Consumer`)
- **Error Handling**: Always handle errors explicitly, use error wrapping
- **Comments**: Add Go doc comments for all exported functions and types

```go
// Good example
package orchestrator

// Publisher defines the interface for publishing messages
type Publisher interface {
    Publish(ctx context.Context, message Message) error
}

// Current project style (uses numeric suffixes)
package controller8

type Controller8OrchestratorM8Interface interface {
    InitOrchestrator() error
    StartOrchestrator(ctx context.Context)
}
```

#### Code Quality Guidelines
1. **Function Length**: Keep functions under 30 lines when possible
2. **Cyclomatic Complexity**: Aim for complexity < 10 per function
3. **Error Handling**: Never ignore errors, always return or log appropriately
4. **Resource Management**: Always use defer for cleanup (connections, files, etc.)

### Testing Strategy

#### Current State
⚠️ **Critical Issue**: No tests currently exist in the codebase

#### Testing Requirements
All new code must include:
1. **Unit Tests**: Test individual functions and methods in isolation
2. **Integration Tests**: Test interactions between components
3. **Mock Dependencies**: Use interfaces to mock external dependencies

#### Testing Structure
```bash
# Create test files alongside source files
pkg/
  amqpM8/
    amqpM8.go
    amqpM8_test.go          # Unit tests
    amqpM8_integration_test.go  # Integration tests
  db8/
    db8_test.go
    db8_domain8_test.go
  orchestratorm8/
    orchestratorm8_test.go
    
# Test utilities and mocks
testdata/
  mocks/
    mock_amqp.go
    mock_db.go
  fixtures/
    test_config.yaml
    sample_domains.json
```

#### Example Test Structure
```go
// pkg/db8/db8_domain8_test.go
package db8

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestDb8Domain8_GetAllEnabled(t *testing.T) {
    tests := []struct {
        name           string
        setupDB        func(*sql.DB)
        expectedCount  int
        expectedError  bool
    }{
        {
            name: "returns enabled domains",
            setupDB: func(db *sql.DB) {
                // Setup test data
            },
            expectedCount: 2,
            expectedError: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Development Commands

#### Essential Commands
```bash
# Format code
go fmt ./...

# Run tests (when implemented)
go test -v -race ./...

# Run with coverage (when tests exist)
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Lint code (requires golangci-lint)
golangci-lint run

# Security scan (requires gosec)
gosec ./...

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o bin/orchestratorm8-linux .
GOOS=windows GOARCH=amd64 go build -o bin/orchestratorm8.exe .
```

#### Recommended Development Tools
```bash
# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
go install github.com/onsi/ginkgo/v2/ginkgo@latest  # For BDD testing
```

## Configuration Management

### Configuration Structure
```yaml
# configs/configuration_template.yaml
APP_ENV: DEV  # DEV, TEST, PROD (DEV and TEST print logs in console)
LOG_LEVEL: "0"  # 0=debug, 1=info, 2=warn, 3=error

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
  asmm8:
    Queue:
      - "cptm8"      # exchange_name
      - "qasmm8"     # queue_name
      - 1            # prefetch count
    Routing-keys:
      - "cptm8.asmm8.#"
    Queue-arguments:
      "x-max-length": 1
      "x-overflow": "reject-publish"
    Consumer:
      - "qasmm8"     # queue name
      - "casmm8"     # consumer name
      - "false"      # autoack

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
    max_connections:     # default: 10
    min_connections:     # default: 2
    max_idle_time:       # default: "1h"
    max_lifetime:        # default: "2h"
    health_check_period: # default: "30m"
    connection_timeout:  # default: "30s"
    retry_attempts:      # default: 3
    retry_delay:         # default: "2s"
```

### Environment Variables
```bash
# Required - Database
export POSTGRESQL_HOSTNAME="localhost"
export POSTGRESQL_DB="cpt_orchestrator"
export POSTGRESQL_USERNAME="cpt_dbuser"
export POSTGRESQL_PASSWORD="secure_database_password"

# Required - RabbitMQ
export RABBITMQ_HOSTNAME="localhost"
export RABBITMQ_USERNAME="orchestrator"
export RABBITMQ_PASSWORD="secure_rabbitmq_password"

# Required - Services
export ASMM8_URL="http://localhost:8000"
export NAABUM8_URL="http://localhost:8001"
export KATANAM8_URL="http://localhost:8002"
export NUM8_URL="http://localhost:8003"

# Optional - Application
export APP_ENV="DEV"           # DEV, TEST, PROD
export LOG_LEVEL="0"           # 0=debug, 1=info, 2=warn, 3=error
```

### Configuration Validation
The application should validate required configuration at startup:
```go
// pkg/configparser/validation.go (to be implemented)
func ValidateConfig(config *Config) error {
    if config.Database.Password == "" {
        return errors.New("database password is required")
    }
    // Additional validation...
}
```

## Database Schema

### Required Tables
```sql
-- Domain management table
CREATE TABLE cptm8domain (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain VARCHAR(255) NOT NULL UNIQUE,
    company VARCHAR(255) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_cptm8domain_enabled ON cptm8domain(enabled);
CREATE INDEX idx_cptm8domain_company ON cptm8domain(company);

-- Sample data
INSERT INTO cptm8domain (domain, company, enabled) VALUES 
('example.com', 'Example Corp', true),
('test-site.org', 'Test Organization', true),
('disabled-site.com', 'Disabled Corp', false);
```

### Database Migrations
Create migration scripts for database changes:
```bash
# migrations/
001_initial_schema.up.sql
001_initial_schema.down.sql
002_add_domain_indexes.up.sql
002_add_domain_indexes.down.sql
```

## Debugging

### Logging Configuration
```yaml
# configuration.yaml - logging section
Logging:
  Level: "debug"  # For development
  Console: true   # Enable console output
  File: true      # Enable file output
  Elasticsearch:
    Enabled: false  # Disable for local development
```

### Debug Mode
```bash
# Run with debug logging
LOG_LEVEL=debug ./bin/orchestratorm8 orchestrator

# Enable Go runtime debugging
GODEBUG=gctrace=1 ./bin/orchestratorm8 orchestrator
```

### Common Issues and Solutions

#### "Connection refused" errors
```bash
# Check services are running
sudo systemctl status postgresql
sudo systemctl status rabbitmq-server

# Check network connectivity
telnet localhost 5432  # PostgreSQL
telnet localhost 5672  # RabbitMQ
```

#### Database connection issues
```bash
# Test database connection
psql -h localhost -U cpt_dbuser -d cpt_orchestrator

# Check database logs
sudo tail -f /var/log/postgresql/postgresql-*.log
```

#### RabbitMQ connection issues
```bash
# Check RabbitMQ status
sudo rabbitmqctl status

# Check RabbitMQ logs
sudo tail -f /var/log/rabbitmq/rabbit@*.log

# Access management interface
# http://localhost:15672 (guest/guest)
```

## Docker Development

### Development Docker Compose
```yaml
# docker-compose.dev.yml
version: '3.8'
services:
  postgres:
    image: postgres:14
    environment:
      POSTGRES_DB: cpt_orchestrator
      POSTGRES_USER: cpt_dbuser
      POSTGRES_PASSWORD: development_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d

  rabbitmq:
    image: rabbitmq:3-management
    environment:
      RABBITMQ_DEFAULT_USER: orchestrator
      RABBITMQ_DEFAULT_PASS: development_password
    ports:
      - "5672:5672"
      - "15672:15672"  # Management UI

  orchestratorm8:
    build: .
    depends_on:
      - postgres
      - rabbitmq
    environment:
      DB_HOST: postgres
      DB_PASSWORD: development_password
      RABBITMQ_HOST: rabbitmq
      RABBITMQ_PASSWORD: development_password
    volumes:
      - ./configuration.yaml:/app/configuration.yaml
      - ./log:/app/log

volumes:
  postgres_data:
```

### Build and Run with Docker
```bash
# Start development environment
docker-compose -f docker-compose.dev.yml up -d

# Build and test
docker-compose -f docker-compose.dev.yml exec orchestratorm8 go test ./...

# View logs
docker-compose -f docker-compose.dev.yml logs -f orchestratorm8
```

## Contributing Guidelines

### Pull Request Process
1. **Create Feature Branch**: `git checkout -b feature/your-feature-name`
2. **Write Tests**: All new code must include comprehensive tests
3. **Update Documentation**: Update relevant documentation files
4. **Security Review**: Ensure no hardcoded secrets or security issues
5. **Performance Check**: Consider performance implications of changes

### Code Review Checklist
- [ ] Tests written and passing
- [ ] No hardcoded credentials or secrets
- [ ] Proper error handling implemented  
- [ ] Documentation updated
- [ ] Go conventions followed
- [ ] Resource cleanup with defer statements
- [ ] Security best practices followed

### Git Workflow
```bash
# Feature development
git checkout -b feature/implement-health-checks
# Make changes, write tests
git add .
git commit -m "Add health check endpoints with tests"
git push origin feature/implement-health-checks
# Create PR

# Keep feature branch updated
git checkout main
git pull origin main
git checkout feature/implement-health-checks  
git rebase main
```

## Performance Considerations

### Development Performance Tips
1. **Use Connection Pooling**: Configure database connection pools appropriately
2. **Avoid N+1 Queries**: Use batch operations for database queries
3. **Memory Management**: Pre-allocate slices when size is known
4. **Goroutine Management**: Don't create unlimited goroutines

### Profiling During Development
```bash
# Build with profiling
go build -ldflags="-X main.enablePprof=true" .

# Run with profiling endpoint
./orchestratorm8 orchestrator &
go tool pprof http://localhost:6060/debug/pprof/profile
```

## IDE Setup

### VS Code Configuration
Create `.vscode/settings.json`:
```json
{
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.testFlags": ["-v", "-race"],
    "go.vetFlags": ["-composites=false"],
    "editor.formatOnSave": true,
    "go.coverOnSave": true
}
```

### Recommended Extensions
- Go (Google)
- GitLens
- Docker
- YAML
- PostgreSQL (Chris Kolkman)

This development guide provides the foundation for contributing to OrchestratorM8. Focus on implementing tests, fixing security issues, and following Go best practices for all new development.