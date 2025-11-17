# Performance Analysis and Optimization Guide

## Current Performance Assessment

### Overall Score: 6/10 (Good Foundation, Room for Optimization)

The OrchestratorM8 system has a solid performance foundation with implemented RabbitMQ connection pooling and proper resource management. Performance is acceptable for small to medium-scale operations. Main optimization opportunities include database polling frequency, caching implementation, and database connection pool configuration.

## Performance-Critical Areas

### 1. Database Operations (High Impact)

#### Current Implementation
**File**: `pkg/controller8/controller8_orchestratorm8.go:106`
```go
// Polling with ticker every 15 minutes
ticker := time.NewTicker(15 * time.Minute)
defer ticker.Stop()

for {
    select {
    case <-ctx.Done():
        return
    case <-ticker.C:
        domains, err := domain8.GetAllEnabled()  // Query enabled domains
        // Process domains...
    }
}
```

**Current Status**:
- ✅ Uses ticker instead of sleep for better timing
- ✅ Context-aware for graceful shutdown
- ⚠️ Still polls database every 15 minutes regardless of changes
- ⚠️ 15-minute interval hardcoded (should be configurable)
- ⚠️ No query result caching implemented

#### Performance Impact
- **CPU**: Unnecessary processing every 15 minutes
- **Memory**: Potential memory leaks from unclosed connections
- **Database Load**: Continuous polling creates constant load
- **Scalability**: Won't scale beyond single instance

### 2. Message Queue Operations (Good)

#### Current Implementation
**File**: `pkg/amqpM8/pooled_amqp.go` (842 lines)
```go
// Advanced connection pooling with:
// - Global pool manager with configurable sizes (max: 10, min: 2)
// - Automatic connection checkout/return
// - Health monitoring (30-minute intervals)
// - Retry logic (3 attempts, 2-second delay)
// - Consumer health tracking
```

**Strengths**:
- ✅ Sophisticated connection pooling (5 files, ~1,800 lines)
- ✅ Automatic reconnection on failures
- ✅ Configurable pool parameters
- ✅ Health check monitoring for consumers
- ✅ Thread-safe operations (sync.RWMutex)

**Potential Optimizations**:
- ⚠️ JSON marshaling still creates objects (could use sync.Pool)
- ⚠️ No batch processing capabilities yet
- ⚠️ Complex architecture may impact debugging

#### Performance Impact
- **Memory**: Well-managed through connection pooling
- **CPU**: Reasonable marshaling overhead
- **Reliability**: ✅ Built-in retry mechanisms

### 3. Concurrent Processing (Medium Impact)

#### Current Issues
**File**: `pkg/orchestratorm8/orchestratorm8.go:173-178`
```go
go func() {
    for {
        checkDBEmpty()  // Runs in goroutine
    }
}()

var c chan bool  // Uninitialized channel
<-c              // Infinite block
```

**Problems**:
- Infinite blocking on uninitialized channel
- No graceful shutdown mechanism
- Two goroutines without coordination
- No error recovery in goroutines

### 4. Resource Management (Good)

#### Current Implementation
**File**: `pkg/controller8/controller8_orchestratorm8.go`
```go
// Connection pooling pattern (WithPooledConnection)
err := amqpM8.WithPooledConnection(func(am8 amqpM8.PooledAmqpInterface) error {
    // Use connection from pool
    // Automatically returned to pool after use
    return nil
})
```

**Strengths**:
- ✅ Connection pooling handles lifecycle automatically
- ✅ Checkout/return pattern prevents leaks
- ✅ Context-aware shutdown (`<-ctx.Done()`)
- ✅ Cleanup of tmp directory on startup (24+ hour threshold)
- ✅ Graceful goroutine termination

**Remaining Optimization**:
- ⚠️ Database connection pooling parameters not explicitly configured
- ⚠️ Could add monitoring for pool utilization metrics

## Detailed Performance Bottlenecks

### Database Layer Bottlenecks

#### 1. Connection Management
```go
// pkg/db8/db8.go - Missing connection pool configuration
func (db *Db8) NewConnection() (*sql.DB, error) {
    // No pool size configuration
    // No connection lifetime management
    // No idle connection limits
}
```

**Recommended Configuration**:
```go
db.SetMaxOpenConns(25)          // Limit concurrent connections
db.SetMaxIdleConns(25)          // Keep connections alive
db.SetConnMaxLifetime(5 * time.Minute)  // Rotate connections
```

#### 2. Query Optimization
```go
// pkg/db8/db8_domain8.go:47-63 - Inefficient slice operations
for rows.Next() {
    domains = append(domains, d)  // No pre-allocation
}
```

**Memory Impact**: Slice grows dynamically, causing multiple memory reallocations

#### 3. Database Polling vs. Event-Driven
Current approach polls database every 15 minutes:
```go
time.Sleep(15 * time.Minute)  // Wasteful polling
```

**Alternative**: PostgreSQL LISTEN/NOTIFY or triggers for real-time updates

### Memory Management Issues

#### 1. Slice Pre-allocation
```go
// Current inefficient approach
var domains []Domain8
for rows.Next() {
    domains = append(domains, domain)  // Causes reallocations
}

// Optimized approach
domains := make([]Domain8, 0, estimatedSize)  // Pre-allocate capacity
```

#### 2. JSON Object Reuse
```go
// Current: Creates new objects every time
messageBody, err := json.Marshal(message)

// Optimized: Use object pools
var messagePool = sync.Pool{
    New: func() interface{} {
        return &bytes.Buffer{}
    },
}
```

### Concurrency Performance Issues

#### 1. Goroutine Management
```go
// Current: Uncontrolled goroutine creation
go func() {
    // Long-running goroutine without context
}()

// Recommended: Context-aware goroutines
ctx, cancel := context.WithCancel(context.Background())
go func(ctx context.Context) {
    select {
    case <-ctx.Done():
        return  // Graceful shutdown
    // ... work
    }
}(ctx)
```

#### 2. Channel Buffering
```go
// Current: Unbuffered channels may cause blocking
notifications := make(chan Notification8)

// Optimized: Buffered channels for better throughput
notifications := make(chan Notification8, 100)
```

## Current Optimization Strategies (Positive Aspects)

### 1. Connection Pooling (RabbitMQ)
- **Advanced pooling**: Global pool manager with health monitoring
- Configurable parameters (max/min connections, timeouts, retries)
- Automatic reconnection and consumer health tracking
- Thread-safe operations with sync.RWMutex

### 2. Structured Logging
- **zerolog** with appropriate log levels (0=debug to 3=error)
- Log rotation with lumberjack (100MB max, 3 backups, 7 days retention)
- Conditional console output based on environment (DEV/TEST)
- File permissions: 0640 for security

### 3. Configuration Management
- Viper with file watching capabilities
- Environment variable substitution for all sensitive data
- Comprehensive pool configuration with defaults
- Externalized configuration (configs/configuration_template.yaml)

### 4. Resource Management
- Context-aware goroutine management
- Proper cleanup with defer statements
- Graceful shutdown handling
- Temporary file cleanup (24-hour threshold)

### 5. Interface Design
- Good separation of concerns
- Testable architecture (ready for unit tests)
- Dependency injection patterns

### 6. Database Transactions
- Proper transaction handling with rollback
- Prepared statements for SQL injection protection

## Performance Optimization Recommendations

### High Priority (Immediate Impact)

#### 1. ✅ Resource Management - ALREADY IMPLEMENTED
The connection pooling system handles resource management automatically:
- Pooled connections are checked out and returned
- Health monitoring ensures connection validity
- Context-based shutdown prevents leaks
- No manual defer statements needed (pool manages lifecycle)

#### 2. Configure Database Connection Pool
```go
// pkg/db8/db8.go
func (db *Db8) NewConnection() (*sql.DB, error) {
    sqlDB, err := sql.Open("postgres", connectionString)
    if err != nil {
        return nil, err
    }
    
    // Configure connection pool
    sqlDB.SetMaxOpenConns(25)
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetConnMaxLifetime(5 * time.Minute)
    
    return sqlDB, nil
}
```

#### 3. Implement Graceful Shutdown
```go
// main.go
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    
    // Handle shutdown signals
    go func() {
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
        <-sigChan
        cancel()
    }()
    
    // Pass context to orchestrator
    orchestrator.Start(ctx)
}
```

#### 4. Pre-allocate Slices
```go
// pkg/db8/db8_domain8.go
func (db *Db8Domain8) GetAllEnabled() ([]model8.Domain8, error) {
    // Get count first for pre-allocation
    var count int
    err := db.db.QueryRow("SELECT COUNT(*) FROM cptm8domain WHERE enabled = true").Scan(&count)
    if err != nil {
        return nil, err
    }
    
    // Pre-allocate slice
    domains := make([]model8.Domain8, 0, count)
    // ... rest of implementation
}
```

### Medium Priority (Performance Improvements)

#### 1. Replace Database Polling with Events
```sql
-- Create trigger for domain changes
CREATE OR REPLACE FUNCTION notify_domain_change()
RETURNS trigger AS $$
BEGIN
    NOTIFY domain_updated, NEW.id::text;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER domain_change_trigger
    AFTER INSERT OR UPDATE ON cptm8domain
    FOR EACH ROW EXECUTE FUNCTION notify_domain_change();
```

```go
// pkg/db8/db8_domain8.go
func (db *Db8Domain8) WatchDomainChanges(ctx context.Context) (<-chan string, error) {
    listener := pq.NewListener(connectionString, 10*time.Second, time.Minute, nil)
    err := listener.Listen("domain_updated")
    if err != nil {
        return nil, err
    }
    
    notifications := make(chan string, 10)
    go func() {
        defer listener.Close()
        for {
            select {
            case notification := <-listener.Notify:
                notifications <- notification.Extra
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return notifications, nil
}
```

#### 2. Implement Caching Layer
```go
// pkg/cache/domain_cache.go
type DomainCache struct {
    cache map[string][]model8.Domain8
    mutex sync.RWMutex
    ttl   time.Duration
}

func (c *DomainCache) GetEnabledDomains() ([]model8.Domain8, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    domains, exists := c.cache["enabled"]
    return domains, exists
}
```

#### 3. Message Processing Optimization
```go
// pkg/amqpM8/message_pool.go
var messageBufferPool = sync.Pool{
    New: func() interface{} {
        return bytes.NewBuffer(make([]byte, 0, 1024))
    },
}

func (a *AmqpM8) PublishMessage(message interface{}) error {
    buf := messageBufferPool.Get().(*bytes.Buffer)
    defer messageBufferPool.Put(buf)
    buf.Reset()
    
    encoder := json.NewEncoder(buf)
    if err := encoder.Encode(message); err != nil {
        return err
    }
    
    return a.channel.Publish(
        exchange, routingKey, false, false,
        amqp.Publishing{Body: buf.Bytes()},
    )
}
```

### Low Priority (Long-term Optimization)

#### 1. Metrics and Monitoring
```go
// pkg/metrics/metrics.go
import "github.com/prometheus/client_golang/prometheus"

var (
    messagePublishDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "message_publish_duration_seconds",
            Help: "Duration of message publishing operations",
        },
        []string{"exchange", "routing_key"},
    )
    
    databaseQueryDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "database_query_duration_seconds", 
            Help: "Duration of database queries",
        },
        []string{"query_type"},
    )
)
```

#### 2. Horizontal Scaling Preparation
```go
// pkg/orchestratorm8/leader_election.go
func (o *Orchestrator8) RunWithLeaderElection(ctx context.Context) error {
    // Implement leader election for multiple instances
    // Use database-based or etcd-based coordination
}
```

## Performance Testing Strategy

### 1. Benchmark Tests
```go
// pkg/amqpM8/amqpM8_bench_test.go
func BenchmarkPublishMessage(b *testing.B) {
    amqp := setupTestAMQP()
    message := createTestMessage()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        amqp.PublishMessage(message)
    }
}

func BenchmarkGetAllEnabled(b *testing.B) {
    db := setupTestDB()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        db.GetAllEnabled()
    }
}
```

### 2. Load Testing
```bash
# Use hey for HTTP load testing (if HTTP endpoints exist)
hey -n 1000 -c 10 http://localhost:8080/health

# Use custom load testing for message queue
go run tools/load_test.go -messages=1000 -concurrency=10
```

### 3. Memory Profiling
```go
// main.go
import _ "net/http/pprof"

func main() {
    if enablePprof {
        go func() {
            log.Println(http.ListenAndServe("localhost:6060", nil))
        }()
    }
    // ... rest of main
}
```

```bash
# Profile memory usage
go tool pprof http://localhost:6060/debug/pprof/heap

# Profile CPU usage
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Check for goroutine leaks
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

## Performance Monitoring

### Key Metrics to Track

#### Application Metrics
- Message publishing rate (messages/second)
- Database query duration (p50, p95, p99)
- Memory usage and GC frequency
- Goroutine count
- Active database connections

#### System Metrics
- CPU utilization
- Memory usage
- Disk I/O for logs and database
- Network traffic for RabbitMQ

#### Business Metrics
- Domain processing rate
- Time from domain update to scan initiation
- Error rates for each service integration

### Monitoring Implementation
```go
// pkg/metrics/collector.go
type MetricsCollector struct {
    messagesPublished prometheus.Counter
    dbConnections     prometheus.Gauge
    goroutineCount    prometheus.Gauge
}

func (m *MetricsCollector) RecordMessagePublished() {
    m.messagesPublished.Inc()
}

func (m *MetricsCollector) UpdateDBConnections(count int) {
    m.dbConnections.Set(float64(count))
}
```

## Performance Targets

### Current State (Measured/Estimated)
- **Throughput**: ~4 domain checks/hour (15-minute polling)
- **Latency**: Up to 15 minutes from domain update to scan initiation
- **Memory**: Stable with connection pooling (managed lifecycle)
- **CPU**: Low utilization, polling-based architecture
- **Connections**: RabbitMQ pooled (max 10), PostgreSQL single connection
- **Resource Management**: Good (connection pooling, context-based shutdown)

### Target Performance (After Optimization)
- **Throughput**: 100+ domain checks/hour (configurable polling or event-driven)
- **Latency**: <5 seconds from domain update to scan initiation (with LISTEN/NOTIFY)
- **Memory**: Stable with caching layer added
- **CPU**: <20% utilization during normal operations
- **Database**: Configured connection pool (max 25, idle 10)
- **Scalability**: Support for multiple orchestrator instances (leader election)

## Implementation Timeline

### Week 1: Critical Fixes
- [x] ✅ Fix resource management - Connection pooling implemented
- [ ] Configure database connection pooling (explicit parameters)
- [x] ✅ Implement graceful shutdown - Context-based shutdown complete
- [ ] Add basic performance metrics (Prometheus integration)

### Week 2: Performance Improvements  
- [ ] Pre-allocate slices in database operations
- [ ] Implement object pooling for JSON marshaling
- [ ] Add caching layer for frequently accessed data
- [ ] Replace database polling with event notifications

### Week 3: Monitoring and Testing
- [ ] Implement comprehensive performance monitoring
- [ ] Create benchmark tests
- [ ] Set up load testing infrastructure
- [ ] Performance profiling and optimization

### Week 4: Advanced Optimizations
- [ ] Horizontal scaling preparation
- [ ] Advanced caching strategies
- [ ] Connection multiplexing for RabbitMQ
- [ ] Performance documentation and runbooks

This performance optimization plan will significantly improve the system's efficiency, scalability, and resource utilization while maintaining reliability and functionality.