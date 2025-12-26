# Production-Grade Payment Orchestration System
## Complete SDLC Guide for Junior Developers

---

## 📋 PROJECT SUMMARY

This document walks through the complete Software Development Life Cycle (SDLC) of building a production-grade payment orchestration backend system using **Clean Architecture** principles and industry best practices.

**Purpose**: Educational guide showing how senior developers think, design, and implement enterprise-grade systems.

**Tech Stack**:
- **Language**: Go 1.21+
- **Framework**: Fiber (HTTP)
- **Database**: PostgreSQL 15+
- **Architecture**: Clean Architecture (Hexagonal/Ports & Adapters)
- **Principles**: SOLID, DDD (Domain-Driven Design)

---

## 🎯 PHASE 1: REQUIREMENTS ANALYSIS

### Why This Phase Matters
Before writing any code, we must understand WHAT we're building and WHY. This prevents costly rewrites and ensures alignment with business goals.

### What We Did

#### 1.1 Functional Requirements (FRs)
Defined **what the system must do**:
- Accept payment requests from partners
- Route payments to providers (Stripe, PayPal, etc.)
- Track payment status (pending → processing → completed)
- Handle refunds within 90 days
- Provide transaction history and reporting

**Key Insight**: FRs directly map to use cases in our code.

#### 1.2 Non-Functional Requirements (NFRs)
Defined **how well the system must perform**:
- **Security**: HTTPS, encryption, PCI-DSS compliance, API key auth
- **Performance**: <200ms response time, 1000 concurrent requests
- **Reliability**: 99.9% uptime, graceful degradation
- **Scalability**: Horizontal scaling, stateless design
- **Maintainability**: Clean code, 80%+ test coverage

**Key Insight**: NFRs drive architectural decisions (caching, connection pooling, etc.)

#### 1.3 Business Rules
Defined **constraints and validations**:
- Min amount: $0.01, Max: $100,000
- Refund window: 90 days
- Max retries: 3 attempts
- Idempotency: 24-hour duplicate detection window

**Key Insight**: Business rules become domain entity validations.

### Deliverable
✅ `docs/REQUIREMENTS.md` - Complete requirements specification

---

## 🏗️ PHASE 2: SYSTEM DESIGN

### Why This Phase Matters
Good architecture makes code maintainable, testable, and scalable. Poor architecture leads to "big ball of mud" that's impossible to change.

### Clean Architecture Layers (Most Important Concept!)

```
┌─────────────────────────────────────────┐
│  Frameworks & Drivers (Outermost)       │  ← External concerns
│  (HTTP, Database, External APIs)        │     Change frequently
│  ▼ Depends on ▼                         │
├─────────────────────────────────────────┤
│  Interface Adapters (Handlers)          │  ← Converts data formats
│  (Controllers, Presenters, Gateways)    │     DTO ↔ Entity
│  ▼ Depends on ▼                         │
├─────────────────────────────────────────┤
│  Use Cases (Application Logic)          │  ← Orchestrates flow
│  (Business workflows)                   │     App-specific rules
│  ▼ Depends on ▼                         │
├─────────────────────────────────────────┤
│  Domain Entities (Core)                 │  ← Business rules
│  (Business objects & rules)             │     Never changes
└─────────────────────────────────────────┘     Pure Go, no deps
```

**The Golden Rule**: Dependencies point INWARD
- Inner layers know nothing about outer layers
- Domain layer has ZERO external dependencies
- Use cases define interfaces; infrastructure implements them

### Why This Matters

**Bad Architecture** (Most juniors do this):
```go
// Handler directly talks to database - WRONG!
func CreateTransaction(c *fiber.Ctx) {
    db.Exec("INSERT INTO transactions...") // Tightly coupled!
}
```
**Problems**:
- Can't test without database
- Can't switch databases easily
- Business logic mixed with infrastructure
- Impossible to maintain

**Good Architecture** (What we built):
```go
// Handler → Use Case → Repository (interface) → Database
func (h *Handler) CreateTransaction(c *fiber.Ctx) {
    // Handler just coordinates
    output, err := h.useCase.Execute(ctx, input)
}

type UseCase struct {
    repo ports.TransactionRepository // Interface!
}

func (uc *UseCase) Execute(input) {
    // Business logic here
    transaction := entities.NewTransaction(...)
    uc.repo.Create(transaction) // Interface call
}
```
**Benefits**:
- Testable with mocks
- Database-agnostic
- Business logic isolated
- Easy to maintain

### Key Design Patterns Applied

#### 1. Repository Pattern
**Problem**: Domain should not know about database
**Solution**: Abstract data access behind interface
```go
// Port (interface) - defined in use case layer
type TransactionRepository interface {
    Create(ctx, *entities.Transaction) error
    GetByID(ctx, uuid.UUID) (*entities.Transaction, error)
}

// Adapter (implementation) - in infrastructure layer
type PostgresTransactionRepo struct {
    db *sql.DB
}
```

#### 2. Dependency Injection
**Problem**: Hard to test, tightly coupled
**Solution**: Inject dependencies through constructors
```go
// Use case receives dependencies as interfaces
func NewCreateTransactionUseCase(
    repo ports.TransactionRepository,  // Interface!
    gateway ports.PaymentGateway,      // Interface!
) *CreateTransactionUseCase {
    return &CreateTransactionUseCase{
        repo: repo,
        gateway: gateway,
    }
}
```

#### 3. Factory Pattern
**Problem**: Complex object creation
**Solution**: Factory methods ensure valid objects
```go
// Can't create invalid transaction
transaction, err := entities.NewTransaction(
    partnerID, idempotencyKey, money, method, provider, email,
)
// All validation happens in factory
```

#### 4. Value Objects
**Problem**: Primitive obsession (using float64 for money is dangerous!)
**Solution**: Encapsulate in immutable value objects
```go
// Money is a value object with validation
type Money struct {
    Amount   float64
    Currency Currency
}

func NewMoney(amount float64, currency string) (Money, error) {
    if amount < 0.01 { return err } // Business rule!
    if amount > 100000 { return err } // Business rule!
    // ...
}
```

### Deliverable
✅ `docs/ARCHITECTURE.md` - Complete system design

---

## 🗄️ PHASE 3: DATABASE DESIGN

### Why This Phase Matters
Database is the source of truth. Good schema design prevents data integrity issues and performance problems.

### Design Principles Applied

#### 1. Normalization (3NF)
- No duplicate data
- Each table has a single responsibility
- Foreign keys maintain referential integrity

#### 2. Indexes for Performance
```sql
-- Composite index for common query pattern
CREATE INDEX idx_transactions_partner_status_created 
    ON transactions(partner_id, status, created_at DESC);

-- Enables fast queries like:
-- SELECT * FROM transactions 
-- WHERE partner_id = ? AND status = ? 
-- ORDER BY created_at DESC;
```

#### 3. Constraints for Data Integrity
```sql
-- Business rules enforced at database level
CONSTRAINT check_amount_positive CHECK (amount >= 0.01),
CONSTRAINT check_amount_max CHECK (amount <= 100000.00),
CONSTRAINT unique_partner_idempotency UNIQUE(partner_id, idempotency_key)
```

#### 4. Triggers for Automation
```sql
-- Auto-update updated_at timestamp
CREATE TRIGGER update_transactions_updated_at 
    BEFORE UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Auto-log transaction changes
CREATE TRIGGER log_transaction_changes_trigger
    AFTER INSERT OR UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION log_transaction_changes();
```

#### 5. Soft Deletes
Never physically delete data (compliance, audit):
```sql
deleted_at TIMESTAMP WITH TIME ZONE

-- Queries filter out deleted rows
WHERE deleted_at IS NULL
```

### Key Tables

**partners** - Stores API clients
- Hashed API keys (never plaintext!)
- Rate limits per partner
- Webhook configuration

**transactions** - Core entity
- Idempotency key (prevents duplicates)
- JSONB metadata (flexible partner data)
- Audit fields (who, when, what)

**refunds** - Linked to transactions
- Validation trigger (prevents over-refunding)
- Maintains refund history

**transaction_events** - Audit trail
- Immutable log of all state changes
- For debugging and compliance

**audit_logs** - Security logs
- All API operations logged
- Compliance requirement (PCI-DSS)

### Performance Optimizations

1. **Indexes on frequent queries**
2. **JSONB for flexible data** (no schema changes needed)
3. **Partitioning preparation** (for massive scale)
4. **Connection pooling** (reuse connections)

### Deliverable
✅ `migrations/000001_init_schema.up.sql` - Database schema
✅ `migrations/000001_init_schema.down.sql` - Rollback migration
✅ `docs/DATABASE_DESIGN.md` - Schema documentation

---

## 💎 PHASE 4: DOMAIN LAYER IMPLEMENTATION

### Why This Phase Matters
Domain layer is the **HEART** of the application. It contains business rules and is completely independent of frameworks.

### What We Built

#### 1. Domain Errors (`internal/domain/errors/`)
Custom error types for business rule violations:
```go
var (
    ErrInvalidAmount = errors.New("invalid transaction amount")
    ErrRefundWindowExpired = errors.New("refund window has expired")
)

type DomainError struct {
    Code    string
    Message string
    Err     error
}
```

**Why**: Explicit error handling, better debugging

#### 2. Value Objects (`internal/domain/valueobjects/`)
Immutable objects that encapsulate validation:

**Money** - Prevents primitive obsession
```go
type Money struct {
    Amount   float64
    Currency Currency
}

// Validation in constructor
func NewMoney(amount, currency) (Money, error) {
    if amount < 0.01 { return err }
    if amount > 100000 { return err }
    // ...
}

// Money-specific operations
func (m Money) Add(other Money) (Money, error)
func (m Money) IsGreaterThan(other Money) bool
```

**PaymentMethod** - Type-safe payment methods
```go
type PaymentMethod string

const (
    PaymentMethodCard PaymentMethod = "card"
    PaymentMethodBankTransfer = "bank_transfer"
)

func NewPaymentMethod(method string) (PaymentMethod, error) {
    // Validates against allowed methods
}
```

**Why Value Objects?**
- **Type Safety**: Can't accidentally use wrong value
- **Validation**: Invalid objects can't exist
- **Operations**: Money.Add() instead of amount1 + amount2
- **Immutability**: Thread-safe, cacheable

#### 3. Domain Entities (`internal/domain/entities/`)
Objects with identity and lifecycle:

**Transaction** - Aggregate Root
```go
type Transaction struct {
    ID             uuid.UUID
    PartnerID      uuid.UUID
    Amount         valueobjects.Money  // Value object!
    Status         TransactionStatus
    // ... more fields
}

// Factory method ensures valid creation
func NewTransaction(...) (*Transaction, error) {
    // All validation here
    if partnerID == uuid.Nil { return err }
    if idempotencyKey == "" { return err }
    // ...
}

// Business logic methods
func (t *Transaction) MarkAsCompleted(providerID string) error {
    // State transition validation
    if t.Status != StatusProcessing {
        return errors.New("can only complete processing transactions")
    }
    // ...
}

func (t *Transaction) IsRefundable() bool {
    // Business rule: 90-day window
    ninetyDaysAgo := time.Now().AddDate(0, 0, -90)
    return t.CreatedAt.After(ninetyDaysAgo)
}
```

**Key Design Decisions**:
1. **No setters** - Use methods that enforce business rules
2. **State transitions** - Invalid transitions rejected
3. **Self-contained logic** - Transaction knows if it's refundable
4. **Factory pattern** - Can't create invalid entities

**Partner** - Another aggregate
- Manages API keys (hashed with bcrypt)
- Validates API keys
- Manages activation state

**Refund** - Entity linked to Transaction
- Validates refund amounts
- Manages refund lifecycle

### SOLID Principles in Domain Layer

**S**ingle Responsibility:
- Transaction only manages transaction logic
- Money only handles monetary operations

**O**pen/Closed:
- Can add new payment methods without changing existing code

**L**iskov Substitution:
- Not applicable (no inheritance in domain)

**I**nterface Segregation:
- Not applicable (interfaces in use case layer)

**D**ependency Inversion:
- Domain has ZERO external dependencies
- 100% pure Go code

### Deliverable
✅ `internal/domain/errors/errors.go`
✅ `internal/domain/valueobjects/money.go`
✅ `internal/domain/valueobjects/payment_method.go`
✅ `internal/domain/entities/transaction.go`
✅ `internal/domain/entities/partner.go`
✅ `internal/domain/entities/refund.go`

---

## 🔄 PHASE 5: USE CASE LAYER IMPLEMENTATION

### Why This Phase Matters
Use cases orchestrate the business workflow. They coordinate between entities, repositories, and external services.

### What We Built

#### 1. Ports (Interfaces) (`internal/usecases/ports/`)
Define contracts for external dependencies:

```go
// Repository interface (Dependency Inversion!)
type TransactionRepository interface {
    Create(ctx, *entities.Transaction) error
    GetByID(ctx, uuid.UUID) (*entities.Transaction, error)
    GetByIdempotencyKey(ctx, uuid.UUID, string) (*entities.Transaction, error)
    // ...
}

// Payment gateway interface
type PaymentGateway interface {
    ProcessPayment(ctx, *entities.Transaction) (providerTxnID string, err error)
    ProcessRefund(ctx, *entities.Refund, *entities.Transaction) (string, error)
    // ...
}
```

**Why Interfaces?**
- Use cases don't know about PostgreSQL or Stripe
- Easy to mock for testing
- Easy to swap implementations (MongoDB, PayPal, etc.)

#### 2. Use Case: Create Transaction

```go
type CreateTransactionUseCase struct {
    transactionRepo ports.TransactionRepository  // Interface!
    partnerRepo     ports.PartnerRepository
    paymentGateway  ports.PaymentGateway
    auditLogger     ports.AuditLogger
}

func (uc *CreateTransactionUseCase) Execute(ctx, input) (*output, error) {
    // Step 1: Validate partner exists and is active
    partner, _ := uc.partnerRepo.GetByID(ctx, input.PartnerID)
    if !partner.IsActive {
        return errors.ErrPartnerInactive
    }
    
    // Step 2: Check idempotency (prevent duplicates)
    existing, _ := uc.transactionRepo.GetByIdempotencyKey(...)
    if existing != nil {
        return existing // Idempotent!
    }
    
    // Step 3: Create value objects (validates amount, currency)
    money, _ := valueobjects.NewMoney(input.Amount, input.Currency)
    
    // Step 4: Create entity (validates business rules)
    transaction, _ := entities.NewTransaction(...)
    
    // Step 5: Persist
    uc.transactionRepo.Create(ctx, transaction)
    
    // Step 6: Audit log
    uc.auditLogger.LogAction(...)
    
    return output
}
```

**Orchestration Steps**:
1. ✅ Validate partner (authentication)
2. ✅ Check idempotency (prevent duplicates)
3. ✅ Validate input (value objects)
4. ✅ Apply business rules (domain entities)
5. ✅ Persist data (repository)
6. ✅ Log audit trail (compliance)

#### 3. Use Case: Process Payment

```go
func (uc *ProcessPaymentUseCase) Execute(ctx, transactionID) error {
    // Get transaction
    txn, _ := uc.transactionRepo.GetByID(ctx, transactionID)
    
    // Validate state transition
    if !txn.IsPending() { return error }
    
    // Mark as processing
    txn.MarkAsProcessing()
    uc.transactionRepo.Update(ctx, txn)
    
    // Call external gateway
    providerTxnID, err := uc.paymentGateway.ProcessPayment(ctx, txn)
    if err != nil {
        // Failed - update transaction
        txn.MarkAsFailed("PAYMENT_FAILED", err.Error())
        uc.transactionRepo.Update(ctx, txn)
        return err
    }
    
    // Success - mark completed
    txn.MarkAsCompleted(providerTxnID)
    uc.transactionRepo.Update(ctx, txn)
    
    // Send webhook (async)
    go uc.notification.SendWebhook(...)
    
    return nil
}
```

**Error Handling Pattern**:
- If payment fails, transaction marked as failed (not thrown away)
- All state changes persisted
- Audit trail maintained

#### 4. Use Case: Refund Transaction

Business rules enforced:
- ✅ Only completed transactions can be refunded
- ✅ Refund within 90-day window
- ✅ Refund amount ≤ transaction amount
- ✅ Total refunds ≤ transaction amount (prevent over-refunding)
- ✅ Currency must match

```go
func (uc *RefundTransactionUseCase) Execute(ctx, input) (*Refund, error) {
    // Get transaction
    txn, _ := uc.transactionRepo.GetByID(...)
    
    // Authorization check
    if txn.PartnerID != input.PartnerID {
        return errors.ErrUnauthorizedOperation
    }
    
    // Business rule: Is refundable?
    if !txn.IsRefundable() {
        return errors.ErrRefundNotAllowed
    }
    
    // Business rule: 90-day window
    if txn.CreatedAt.Before(ninetyDaysAgo) {
        return errors.ErrRefundWindowExpired
    }
    
    // Business rule: Check total refunded amount
    totalRefunded, _ := uc.refundRepo.GetTotalRefundedAmount(...)
    if totalRefunded + input.Amount > txn.Amount.Amount {
        return errors.ErrRefundAmountExceeded
    }
    
    // Create refund entity
    refund, _ := entities.NewRefund(...)
    
    // Process through gateway
    providerRefundID, _ := uc.paymentGateway.ProcessRefund(...)
    
    // Update states
    refund.MarkAsCompleted(providerRefundID)
    txn.MarkAsRefunded(isPartial)
    
    return refund
}
```

### Key Patterns in Use Cases

**1. Single Responsibility**: Each use case does ONE thing
**2. Dependency Injection**: All dependencies are interfaces
**3. Transaction Script**: Step-by-step workflow
**4. Error Handling**: Explicit error returns
**5. Audit Logging**: All operations logged
**6. Authorization**: Partner can only access their own data

### Deliverable
✅ `internal/usecases/ports/repositories.go` - Interface definitions
✅ `internal/usecases/transaction/create_transaction.go`
✅ `internal/usecases/transaction/process_payment.go`
✅ `internal/usecases/transaction/refund_transaction.go`

---

## 🎓 KEY LEARNINGS FOR JUNIOR DEVELOPERS

### 1. Clean Architecture Benefits

**Traditional Approach** (Most juniors):
```
Controller → Database
```
- Tightly coupled
- Can't test without database
- Business logic scattered
- Hard to change

**Clean Architecture** (What we built):
```
HTTP Handler → Use Case → Domain Entity
                ↓
          Repository Interface → Database
```
- Loosely coupled
- Easy to test (mock interfaces)
- Business logic centralized
- Easy to swap components

### 2. Why So Many Layers?

**Junior Question**: "Why not just write SQL in the handler?"

**Answer**:
- **Testability**: Can test business logic without database
- **Flexibility**: Can swap PostgreSQL for MongoDB without changing business logic
- **Maintainability**: Each layer has clear responsibility
- **Reusability**: Use cases can be called from HTTP, gRPC, CLI, etc.

### 3. Domain-Driven Design (DDD)

**Value Objects** prevent bugs:
```go
// Bad: Primitive obsession
amount := 100.50
currency := "USD"
// Easy to mix up, forget to validate

// Good: Value object
money, err := valueobjects.NewMoney(100.50, "USD")
// Validated, type-safe, operations built-in
```

**Entities** encapsulate business rules:
```go
// Bad: Business logic in use case
if transaction.Status == "processing" {
    transaction.Status = "completed"
}

// Good: Business logic in entity
err := transaction.MarkAsCompleted(providerID)
// Validates state transition, prevents invalid states
```

### 4. SOLID Principles in Practice

**Single Responsibility**:
- Transaction entity: manages transaction state
- CreateTransactionUseCase: orchestrates creation
- TransactionRepository: handles persistence

**Dependency Inversion**:
- Use cases depend on `TransactionRepository` interface
- PostgreSQL implementation is a detail
- Can easily switch to MongoDB

### 5. Security by Design

- API keys hashed with bcrypt
- SQL injection prevented (parameterized queries)
- Authorization checks in every use case
- Audit logging for compliance
- Soft deletes (never lose data)
- Rate limiting (prevent abuse)

### 6. Error Handling Strategy

**Domain Errors**: Business rule violations
```go
ErrRefundWindowExpired
ErrAmountAboveMaximum
```

**Application Errors**: Operation failures
```go
"failed to get transaction: %w"
```

**Always wrap errors** for context:
```go
return fmt.Errorf("failed to create transaction: %w", err)
```

---

## 📚 REMAINING PHASES (Not Implemented Yet)

### Phase 6: Adapter Layer
- HTTP handlers (Fiber)
- Middleware (auth, logging, rate limit)
- DTOs (request/response models)
- PostgreSQL repository implementation

### Phase 7: Infrastructure Layer
- Payment gateway integrations (Stripe, PayPal)
- Email/webhook notifications
- Redis caching
- Configuration management

### Phase 8: Testing
- Unit tests (domain, use cases)
- Integration tests (database)
- E2E tests (full API flow)
- Test coverage: 80%+

### Phase 9: Deployment
- Dockerfile & docker-compose
- Kubernetes manifests
- CI/CD pipeline
- Monitoring & alerting

---

## 🏆 PRODUCTION-GRADE CHECKLIST

✅ **Requirements**: Documented FRs, NFRs, business rules
✅ **Architecture**: Clean Architecture, SOLID principles
✅ **Database**: Normalized schema, indexes, triggers, migrations
✅ **Domain**: Value objects, entities with business logic
✅ **Use Cases**: Orchestration logic, error handling
✅ **Security**: Encryption, hashing, authorization
✅ **Audit**: Complete audit trail
✅ **Error Handling**: Explicit, contextual errors
✅ **Idempotency**: Duplicate prevention
✅ **State Machines**: Valid state transitions only
✅ **Documentation**: Every design decision explained

❌ **Not Yet Done** (Next Steps):
- HTTP API implementation
- Payment provider integrations
- Comprehensive testing
- Docker deployment
- Monitoring/observability

---

## 💡 HOW TO LEARN FROM THIS CODE

### For Junior Developers:

1. **Read in Order**:
   - docs/REQUIREMENTS.md
   - docs/ARCHITECTURE.md
   - docs/DATABASE_DESIGN.md
   - internal/domain/entities/
   - internal/usecases/

2. **Understand WHY**:
   - Why value objects instead of primitives?
   - Why interfaces for repositories?
   - Why separate layers?

3. **Compare to Your Code**:
   - How do you handle money? (float64 vs Money value object)
   - How do you validate? (in handler vs in entity)
   - How do you test? (with database vs mocked interfaces)

4. **Practice**:
   - Implement the adapter layer
   - Write unit tests for use cases
   - Add a new payment provider

### Key Takeaways:

1. **Think in Layers**: External → Adapter → Use Case → Domain
2. **Validate Early**: Value objects & entities enforce rules
3. **Use Interfaces**: Makes code testable and flexible
4. **Explicit Errors**: Don't hide errors, handle them properly
5. **Business Logic in Domain**: Not in handlers or database
6. **Test Without Infrastructure**: Mock interfaces in tests

---

## 📖 FURTHER READING

- **Clean Architecture** by Robert C. Martin
- **Domain-Driven Design** by Eric Evans
- **Implementing Domain-Driven Design** by Vaughn Vernon
- **The Pragmatic Programmer** by Hunt & Thomas

---

## 🎯 CONCLUSION

This is not just code; it's a **learning resource** showing professional software engineering practices. Every decision was intentional, every pattern has a purpose.

**Junior developers**: Study this to understand how senior developers think.
**Mid-level developers**: Compare to your approach and learn new patterns.
**Senior developers**: Review and provide feedback for improvements.

Remember: **Good code is easy to understand, easy to test, and easy to change.**

---

**Built with ❤️ for educational purposes**
