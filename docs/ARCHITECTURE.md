# OrchestratorM8 System Architecture

## Overview

**OrchestratorM8** is a message queue orchestration service built in Go that acts as a central coordinator for a distributed cybersecurity platform called "CPT" (Continuous Penetration Testing). The system orchestrates multiple security scanning services through RabbitMQ message queues, managing scan workflows and coordinating between various security analysis tools.

## Technology Stack

- **Language**: Go 1.21.5
- **Web Framework**: Gin v1.10.1
- **Message Queue**: RabbitMQ (AMQP 0.9.1) via rabbitmq/amqp091-go v1.10.0
- **Database**: PostgreSQL (lib/pq v1.10.9)
- **CLI Framework**: Cobra v1.8.1
- **Configuration**: Viper v1.19.0 (YAML-based)
- **Logging**: Zerolog v1.33.0 with Elasticsearch integration
- **Log Rotation**: Lumberjack v2.2.1
- **Containerization**: Docker (multi-stage builds with Alpine 3.20)
- **UUID**: GOFRS UUID v5.3.1

## System Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   PostgreSQL    │    │   OrchestratorM8 │    │    RabbitMQ     │
│   Database      │◄──►│                  │◄──►│  Message Queue  │
│                 │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
                   ┌─────────────────────────────┐
                   │     Security Services       │
                   │                             │
                   │ ┌─────────┐ ┌─────────────┐ │
                   │ │ asmm8   │ │   naabum8   │ │
                   │ │:8000    │ │   :8001     │ │
                   │ └─────────┘ └─────────────┘ │
                   │                             │
                   │ ┌─────────┐ ┌─────────────┐ │
                   │ │katanam8 │ │    num8     │ │
                   │ │:8002    │ │   :8003     │ │
                   │ └─────────┘ └─────────────┘ │
                   │                             │
                   │    ┌─────────────────┐      │
                   │    │  reportingm8    │      │
                   │    │                 │      │
                   │    └─────────────────┘      │
                   └─────────────────────────────┘
```

## Core Components

### 1. Orchestrator Core (`pkg/controller8/`)
**Purpose**: Central coordinator that manages message queue flows

**Key Responsibilities**:
- Initialize RabbitMQ exchanges and queues
- Monitor database for domains to scan every 15 minutes
- Publish scan initiation messages to appropriate services
- Coordinate workflow between multiple security services
- Provide health check and readiness endpoints

**Key Files**:
- `controller8_orchestratorm8.go` - Main orchestration logic (267 lines)
- `controller8_orchestratorm8_interface.go` - Interface definitions

### 2. Message Queue Layer (`pkg/amqpM8/`)
**Purpose**: Advanced connection pooling and RabbitMQ operations

**Features**:
- **Connection Pooling**: Global pool manager with configurable pool sizes (default: max 10, min 2)
- **Health Monitoring**: Consumer health tracking with automatic reconnection
- **Shared State**: Thread-safe state management across pooled connections
- **Graceful Shutdown**: Context-aware consumer shutdown
- Queue configuration with max-length and overflow policies
- Support for multiple binding keys per queue

**Key Files**:
- `pooled_amqp.go` - Main pooled AMQP implementation (842 lines)
- `connection_pool.go` - Connection pool management (406 lines)
- `pool_manager.go` - Global pool manager (190 lines)
- `shared_state.go` - Shared state across connections (241 lines)
- `initialization.go` - Pool initialization from configuration (98 lines)

### 3. Database Layer (`pkg/db8/`)
**Purpose**: PostgreSQL integration for domain and scan target management

**Entities**:
- Domain objects with company associations
- Scan configurations and scheduling information

**Operations**:
- CRUD operations for scanning targets
- Query enabled domains for processing
- Transaction management with rollback support

**Key Files**:
- `db8.go` - Database connection and utilities
- `db8_domain8.go` - Domain-specific database operations
- `db8_domain8_interface.go` - Interface definitions

### 4. Configuration Management (`pkg/configparser/`)
**Purpose**: YAML-based configuration with runtime updates

**Features**:
- File watcher for configuration changes
- Environment-specific settings (DEV/TEST/PROD)
- Structured configuration validation

### 5. Data Models (`pkg/model8/`)
**Purpose**: Domain models and message structures

**Models**:
- `domain8.go` - Scanning target domain models
- `amqpm8message8.go` - Message queue message structures
- `notification8.go` - Event notification models

### 6. Logging System (`pkg/log8/`)
**Purpose**: Centralized structured logging

**Features**:
- Zerolog integration with multiple outputs
- Log rotation with configurable size limits
- Elasticsearch integration for log aggregation
- Environment-specific log levels

## Message Queue Architecture

### Exchanges
1. **cptm8** (topic) - Main security scanning coordination
2. **scheduler** (topic) - Scheduled operations and timing
3. **notification** (topic) - Event notifications and alerts

### Queue Configuration
- **Prefetch Count**: 1 (ensures load balancing)
- **Max Length**: 1 (prevents queue overflow)
- **Overflow Policy**: "reject-publish" (prevents memory issues)
- **Acknowledgment**: Manual (autoack: false) for reliability

### Routing Strategy
Messages are routed using topic-based routing keys that correspond to different security services:
- `asmm8.*` - Asset Management routing
- `naabum8.*` - Network Analysis routing  
- `katanam8.*` - Vulnerability Scanning routing
- `num8.*` - Numerical Analysis routing
- `reportingm8.*` - Report Generation routing

## Service Orchestration

### Security Services Managed
1. **asmm8** (port 8000) - Asset Management and Discovery
2. **naabum8** (port 8001) - Network Analysis and Mapping
3. **katanam8** (port 8002) - Vulnerability Scanning and Assessment
4. **num8** (port 8003) - Numerical Analysis and Statistics
5. **reportingm8** - Report Generation and Aggregation

### Workflow Process
1. **Domain Monitoring**: Every 15 minutes, query database for enabled domains
2. **Message Generation**: Create scan initiation messages for each domain
3. **Service Routing**: Route messages to appropriate security services via RabbitMQ
4. **Coordination**: Monitor and coordinate between services
5. **Result Processing**: Handle scan results and status updates

## Design Patterns

### 1. Interface-Driven Design
- Clean separation through interfaces for all major components
- Dependency injection pattern for testability
- Mock-friendly architecture for unit testing

### 2. Event-Driven Architecture
- RabbitMQ-based asynchronous message passing
- Topic exchanges for flexible message routing
- Decoupled service communication

### 3. Configuration-as-Code
- Externalized configuration via YAML
- Environment-specific configuration management
- Hot-reload capability for runtime configuration updates

### 4. Concurrent Processing
- Goroutines for database monitoring (15-minute intervals)
- Concurrent message publishing coordination
- Asynchronous consumer handling

### 5. Domain-Driven Design
- Clear domain models (Domain8, Notification8)
- Business logic encapsulation
- Separation of concerns between layers

## Deployment Architecture

### Docker Configuration
- **Multi-stage build**: Separate build and runtime containers
- **Build Environment**: Go 1.21+ with full toolchain
- **Runtime Environment**: Alpine Linux for minimal footprint
- **Entry Point**: CLI help command for container introspection

### Configuration Management
- YAML-based configuration files
- Environment variable override support
- File watching for configuration hot-reload
- Development vs. production configuration separation

## Key System Behaviors

1. **Periodic Scan Orchestration**: Automated triggering of security scans every 15 minutes
2. **Service Coordination**: Intelligent routing of scan requests to appropriate services
3. **Database-Driven Workflow**: Scan targets managed through PostgreSQL database
4. **Centralized Logging**: Structured logging with Elasticsearch integration for monitoring
5. **Graceful Resource Management**: Proper connection lifecycle management

## Scalability Considerations

### Current Architecture Strengths
- **Connection Pooling**: RabbitMQ connection pool supports multiple concurrent connections
- **Context-Aware Shutdown**: Graceful handling of SIGINT/SIGTERM signals
- **Health Monitoring**: Built-in consumer health tracking and reconnection
- **Resource Management**: Proper cleanup and connection lifecycle management

### Current Architecture Limitations
- Single orchestrator instance (no horizontal scaling)
- Database polling every 15 minutes (not event-driven)
- No load balancing between multiple orchestrator instances
- Database connection pooling parameters not explicitly configured

### Future Scaling Opportunities
- Implement leader election for multiple orchestrator instances
- Move to event-driven architecture with PostgreSQL LISTEN/NOTIFY
- Add horizontal scaling with consistent hashing for domain distribution
- Implement caching layer for frequently accessed domain data
- Configure database connection pooling (SetMaxOpenConns, SetMaxIdleConns)

## Integration Points

### External Systems
- **PostgreSQL Database**: Primary data store for domains and configuration
- **RabbitMQ**: Message queue for service coordination
- **Security Services**: Multiple microservices for different security functions
- **Elasticsearch**: Log aggregation and search (optional)

### API Boundaries
- Database layer through SQL interfaces
- Message queue through AMQP protocol
- Internal service communication through Go interfaces
- External service communication through RabbitMQ messaging

This architecture provides a solid foundation for orchestrating distributed security scanning operations with proper separation of concerns, event-driven communication, and scalable design patterns.