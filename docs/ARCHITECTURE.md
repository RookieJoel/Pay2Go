# System Design - Payment Orchestration Platform

## 1. CLEAN ARCHITECTURE LAYERS

```
┌─────────────────────────────────────────────────────────────┐
│                     EXTERNAL INTERFACES                      │
│  (HTTP Handlers, gRPC, Message Queue Consumers, CLI)        │
│                    adapter/ (Delivery Layer)                 │
└────────────────────┬────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│                     USE CASES / BUSINESS LOGIC               │
│        (Application-specific business rules)                 │
│                    usecases/ (Service Layer)                 │
└────────────────────┬────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│                    DOMAIN ENTITIES                           │
│        (Enterprise business rules & core models)             │
│                    entities/ (Domain Layer)                  │
└─────────────────────────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│              INFRASTRUCTURE / FRAMEWORKS                     │
│   (Database, External APIs, File System, External Services) │
│    repositories/, pkg/payment-providers/, pkg/cache/         │
└─────────────────────────────────────────────────────────────┘
```

### Dependency Rule (Critical!)
- **Inner layers cannot depend on outer layers**
- **Dependencies point inward** (toward domain)
- **Domain entities are independent** of frameworks/databases
- **Use cases define interfaces**, infrastructure implements them

## 2. DIRECTORY STRUCTURE (Production-Grade)

```
Pay2Go/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/                     # Private application code
│   ├── domain/                   # CORE DOMAIN (innermost layer)
│   │   ├── entities/
│   │   │   ├── transaction.go   # Transaction aggregate
│   │   │   ├── partner.go       # Partner entity
│   │   │   ├── payment.go       # Payment value object
│   │   │   └── refund.go        # Refund entity
│   │   ├── valueobjects/
│   │   │   ├── money.go         # Money value object
│   │   │   ├── currency.go      # Currency enum
│   │   │   └── payment_method.go
│   │   └── errors/
│   │       └── domain_errors.go # Domain-specific errors
│   │
│   ├── usecases/                # APPLICATION BUSINESS LOGIC
│   │   ├── transaction/
│   │   │   ├── create_transaction.go
│   │   │   ├── get_transaction.go
│   │   │   ├── process_payment.go
│   │   │   └── refund_transaction.go
│   │   ├── partner/
│   │   │   ├── authenticate.go
│   │   │   └── get_partner.go
│   │   └── ports/               # Interfaces (Dependency Inversion)
│   │       ├── repositories.go  # Repository interfaces
│   │       ├── payment_gateway.go # Payment provider interface
│   │       └── notification.go  # Notification interface
│   │
│   ├── adapters/                # INTERFACE ADAPTERS
│   │   ├── http/
│   │   │   ├── handlers/
│   │   │   │   ├── transaction_handler.go
│   │   │   │   ├── webhook_handler.go
│   │   │   │   └── health_handler.go
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go      # API key authentication
│   │   │   │   ├── rate_limit.go
│   │   │   │   ├── logger.go
│   │   │   │   ├── recovery.go  # Panic recovery
│   │   │   │   ├── cors.go
│   │   │   │   └── request_id.go
│   │   │   ├── dto/             # Data Transfer Objects
│   │   │   │   ├── request.go
│   │   │   │   └── response.go
│   │   │   └── routes.go
│   │   └── persistence/
│   │       ├── postgres/
│   │       │   ├── transaction_repository.go
│   │       │   ├── partner_repository.go
│   │       │   └── migrations/
│   │       └── cache/
│   │           └── redis_cache.go
│   │
│   └── infrastructure/          # EXTERNAL CONCERNS
│       ├── payment/
│       │   ├── stripe_gateway.go
│       │   ├── paypal_gateway.go
│       │   └── factory.go       # Factory pattern for providers
│       ├── notification/
│       │   └── email_service.go
│       ├── config/
│       │   └── config.go        # Configuration management
│       ├── logger/
│       │   └── logger.go        # Structured logging
│       └── database/
│           └── postgres.go      # DB connection pooling
│
├── pkg/                         # PUBLIC LIBRARIES (reusable)
│   ├── encryption/
│   │   └── aes.go              # Encryption utilities
│   ├── validator/
│   │   └── validator.go        # Input validation
│   ├── middleware/
│   │   └── security.go         # Security headers
│   └── errors/
│       └── app_errors.go       # Application error types
│
├── tests/
│   ├── unit/                   # Unit tests (per layer)
│   ├── integration/            # Integration tests
│   └── e2e/                    # End-to-end tests
│
├── migrations/                 # Database migrations
│   ├── 000001_init_schema.up.sql
│   └── 000001_init_schema.down.sql
│
├── deployments/
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   └── k8s/                   # Kubernetes manifests (future)
│
├── docs/
│   ├── REQUIREMENTS.md
│   ├── ARCHITECTURE.md
│   ├── API.md                 # API documentation
│   └── DEPLOYMENT.md
│
├── scripts/
│   ├── migrate.sh
│   └── seed.sh
│
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
├── Makefile                   # Build automation
└── README.md
```

## 3. COMPONENT INTERACTION FLOW

### Payment Creation Flow:
```
HTTP Request → Middleware Chain → Handler → Use Case → Repository → Database
     ↓              ↓                ↓          ↓           ↓
  Validation   Auth/Logging      DTO→Entity   Domain     SQL Query
                                              Validation
```

### Detailed Flow:
1. **HTTP Handler** (adapters/http/handlers)
   - Receives HTTP request
   - Validates request format (JSON)
   - Converts to DTO

2. **Middleware** (adapters/http/middleware)
   - Authentication (API key validation)
   - Rate limiting
   - Request ID generation
   - Logging

3. **Use Case** (usecases/)
   - Converts DTO to Domain Entity
   - Applies business rules
   - Validates domain constraints
   - Coordinates repositories and external services

4. **Repository** (adapters/persistence)
   - Implements port interfaces
   - Handles database operations
   - Converts between domain entities and DB models

5. **Payment Gateway** (infrastructure/payment)
   - Implements payment provider interface
   - Handles external API calls
   - Circuit breaker pattern
   - Retry logic

## 4. KEY DESIGN PATTERNS

### 4.1 Repository Pattern
- Abstract data access layer
- Domain doesn't know about database
- Easier testing with mocks

### 4.2 Factory Pattern
- Payment gateway factory
- Creates appropriate provider based on config
- Enables strategy pattern

### 4.3 Strategy Pattern
- Different payment methods
- Different refund strategies
- Different notification channels

### 4.4 Circuit Breaker
- Prevents cascading failures
- Protects against slow/failing external services
- Auto-recovery mechanism

### 4.5 Unit of Work
- Atomic transactions
- Multiple repository operations
- Rollback on failure

### 4.6 Dependency Injection
- Constructor injection
- Interface-based dependencies
- Testability and flexibility

## 5. DATA FLOW DIAGRAMS

### Transaction Creation:
```
Partner → API Gateway → Rate Limiter → Auth Middleware
                                             ↓
                                    Transaction Handler
                                             ↓
                                 Create Transaction UseCase
                                             ↓
                        ┌────────────────────┴────────────────────┐
                        ↓                                         ↓
              Transaction Repository                    Payment Gateway
                        ↓                                         ↓
                    Database                              Stripe/PayPal
                        ↓                                         ↓
                    Save Record  ←──────────────  Process Payment
                        ↓
                  Return Response
```

## 6. ERROR HANDLING STRATEGY

```go
// Domain Errors (innermost)
type DomainError struct {
    Code    string
    Message string
}

// Application Errors (use case layer)
type AppError struct {
    DomainError
    Operation string
    Err       error
}

// HTTP Errors (adapter layer)
type HTTPError struct {
    StatusCode int
    AppError
}
```

## 7. SECURITY DESIGN

### Defense in Depth:
1. **Network**: HTTPS only, TLS 1.3
2. **API Gateway**: Rate limiting, IP whitelisting
3. **Authentication**: API keys (hashed with bcrypt)
4. **Authorization**: Partner-specific resource access
5. **Input Validation**: Schema validation, sanitization
6. **Data Protection**: Encryption at rest (AES-256)
7. **Audit**: All operations logged with correlation ID
8. **Secrets Management**: Environment variables, never hardcoded

## 8. SCALABILITY DESIGN

### Horizontal Scaling:
- **Stateless API servers** (no session state)
- **Connection pooling** (database connections)
- **Database read replicas** (read/write separation)
- **Caching layer** (Redis for hot data)
- **Message queue** (async processing)
- **Load balancer** (distribute traffic)

### Database Optimization:
- **Indexes** on frequently queried columns
- **Partitioning** for large transaction tables
- **Archival strategy** for old transactions
- **Query optimization** with EXPLAIN ANALYZE

## 9. OBSERVABILITY DESIGN

### Logging Strategy:
```json
{
  "timestamp": "2025-12-26T10:00:00Z",
  "level": "info",
  "service": "pay2go",
  "request_id": "uuid",
  "partner_id": "partner-123",
  "transaction_id": "txn-456",
  "operation": "create_transaction",
  "duration_ms": 150,
  "status": "success"
}
```

### Metrics to Track:
- Request rate (requests/second)
- Error rate (errors/total requests)
- Latency (p50, p95, p99)
- Payment success rate
- Provider-specific metrics

### Health Checks:
- `/health/live` - Liveness probe
- `/health/ready` - Readiness probe
- Database connectivity
- External service status
