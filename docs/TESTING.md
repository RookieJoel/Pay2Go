# Testing Guide

## Pay2Go - Comprehensive Testing Guide

This guide covers all testing strategies for the Pay2Go payment orchestration system.

---

## Testing Strategy

### Test Pyramid

```
        /\
       /  \
      / UI \
     /______\
    /        \
   /Integration\
  /____________\
 /              \
/  Unit Tests    \
```

- **Unit Tests** (70%): Test individual components in isolation
- **Integration Tests** (20%): Test component interactions
- **End-to-End Tests** (10%): Test complete user flows

---

## Running Tests

### All Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test -v ./internal/domain/entities/...

# Run specific test
go test -v -run TestNewTransaction_Success ./tests/unit/domain/
```

### Test Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# Coverage by package
go test -cover ./...
```

---

## Unit Tests

Unit tests validate individual components in isolation.

### Domain Layer Tests

Located in: `tests/unit/domain/`

#### Entity Tests (`transaction_test.go`)

```go
func TestNewTransaction_Success(t *testing.T) {
    // Arrange
    partnerID := uuid.New()
    amount, _ := valueobjects.NewMoney(10000, "USD")
    
    // Act
    transaction, err := entities.NewTransaction(
        partnerID,
        amount,
        valueobjects.PaymentMethodCreditCard,
        valueobjects.PaymentProviderStripe,
        "Test transaction",
        "test-key-123",
        nil,
    )
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, transaction)
    assert.Equal(t, "pending", transaction.Status)
}
```

**What to Test**:
- ✅ Factory methods create valid entities
- ✅ Business rule validation (amount limits, refund windows)
- ✅ State transitions (pending → completed → refunded)
- ✅ Invariant enforcement (can't complete failed transaction)
- ✅ Error handling for invalid inputs

#### Value Object Tests (`money_test.go`)

```go
func TestMoney_Add(t *testing.T) {
    // Table-driven tests
    tests := []struct {
        name    string
        money1  *Money
        money2  *Money
        want    int64
        wantErr bool
    }{
        {
            name:    "Add same currency",
            money1:  mustNewMoney(10000, "USD"),
            money2:  mustNewMoney(5000, "USD"),
            want:    15000,
            wantErr: false,
        },
        {
            name:    "Add different currencies",
            money1:  mustNewMoney(10000, "USD"),
            money2:  mustNewMoney(5000, "EUR"),
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := tt.money1.Add(tt.money2)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, result.Amount)
            }
        })
    }
}
```

**What to Test**:
- ✅ Validation rules (min/max amounts, currency codes)
- ✅ Immutability (operations return new instances)
- ✅ Equality comparison
- ✅ Arithmetic operations (add, subtract)
- ✅ Currency mismatch handling

### Use Case Tests

Located in: `tests/unit/usecases/`

```go
// Example: create_transaction_test.go
func TestCreateTransactionUseCase_Success(t *testing.T) {
    // Arrange - Setup mocks
    mockTxRepo := new(MockTransactionRepository)
    mockPartnerRepo := new(MockPartnerRepository)
    mockGatewayFactory := func(provider string) PaymentGateway {
        return new(MockPaymentGateway)
    }
    
    useCase := NewCreateTransactionUseCase(
        mockTxRepo,
        mockPartnerRepo,
        mockGatewayFactory,
    )
    
    // Setup expectations
    mockPartnerRepo.On("GetByID", mock.Anything, partnerID).
        Return(partner, nil)
    mockTxRepo.On("GetByIdempotencyKey", mock.Anything, idempotencyKey).
        Return(nil, domain.ErrNotFound)
    mockTxRepo.On("Create", mock.Anything, mock.Anything).
        Return(nil)
    
    // Act
    result, err := useCase.Execute(ctx, input)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    mockTxRepo.AssertExpectations(t)
    mockPartnerRepo.AssertExpectations(t)
}
```

**What to Test**:
- ✅ Happy path (successful execution)
- ✅ Idempotency (duplicate requests)
- ✅ Authorization (partner ownership)
- ✅ Business rule enforcement
- ✅ Error propagation from repositories
- ✅ Transaction orchestration

### Repository Tests

For repository tests, use test database or mocks.

```go
func TestTransactionRepository_Create(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()
    
    repo := NewTransactionRepository(db)
    transaction := createTestTransaction(t)
    
    // Act
    err := repo.Create(context.Background(), transaction)
    
    // Assert
    assert.NoError(t, err)
    
    // Verify in database
    var count int
    db.QueryRow("SELECT COUNT(*) FROM transactions WHERE id = $1", 
        transaction.ID).Scan(&count)
    assert.Equal(t, 1, count)
}
```

---

## Integration Tests

Integration tests verify component interactions.

Located in: `tests/integration/`

### Database Integration Tests

```go
func TestTransactionFlow_Integration(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    txRepo := postgres.NewTransactionRepository(db)
    partnerRepo := postgres.NewPartnerRepository(db)
    
    // Create test partner
    partner := createTestPartner(t)
    err := partnerRepo.Create(context.Background(), partner)
    require.NoError(t, err)
    
    // Create transaction
    transaction := createTestTransaction(t, partner.ID)
    err = txRepo.Create(context.Background(), transaction)
    require.NoError(t, err)
    
    // Retrieve transaction
    retrieved, err := txRepo.GetByID(context.Background(), transaction.ID)
    require.NoError(t, err)
    
    // Assert
    assert.Equal(t, transaction.ID, retrieved.ID)
    assert.Equal(t, transaction.Amount.Amount, retrieved.Amount.Amount)
}
```

### API Integration Tests

```go
func TestCreateTransactionAPI_Integration(t *testing.T) {
    // Setup test server
    app := setupTestApp(t)
    
    // Create test partner with API key
    partner, apiKey := createTestPartnerWithAPIKey(t)
    
    // Prepare request
    reqBody := map[string]interface{}{
        "amount":          10000,
        "currency":        "USD",
        "payment_method":  "credit_card",
        "payment_provider": "stripe",
        "description":     "Test payment",
        "idempotency_key": uuid.New().String(),
    }
    
    // Make request
    req := httptest.NewRequest("POST", "/api/v1/transactions", 
        toJSONReader(reqBody))
    req.Header.Set("Authorization", "Bearer "+apiKey)
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := app.Test(req)
    require.NoError(t, err)
    
    // Assert
    assert.Equal(t, 201, resp.StatusCode)
    
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    assert.Equal(t, "pending", result["status"])
}
```

---

## End-to-End Tests

E2E tests verify complete user workflows.

### Payment Flow E2E Test

```go
func TestCompletePaymentFlow_E2E(t *testing.T) {
    // Start test server
    app := startTestServer(t)
    defer app.Shutdown()
    
    baseURL := "http://localhost:8080"
    client := &http.Client{}
    
    // 1. Create partner (setup)
    partner, apiKey := createPartnerViaAPI(t, baseURL)
    
    // 2. Create transaction
    txResp := createTransaction(t, client, baseURL, apiKey, TransactionRequest{
        Amount:          10000,
        Currency:        "USD",
        PaymentMethod:   "credit_card",
        PaymentProvider: "stripe",
        Description:     "Test order",
        IdempotencyKey:  uuid.New().String(),
    })
    
    assert.Equal(t, "pending", txResp.Status)
    txID := txResp.ID
    
    // 3. Process payment
    processResp := processPayment(t, client, baseURL, apiKey, txID)
    assert.Equal(t, "completed", processResp.Status)
    
    // 4. Verify transaction
    tx := getTransaction(t, client, baseURL, apiKey, txID)
    assert.Equal(t, "completed", tx.Status)
    assert.NotEmpty(t, tx.ProviderTransactionID)
    
    // 5. Refund transaction
    refundResp := refundTransaction(t, client, baseURL, apiKey, txID, RefundRequest{
        Amount: 5000,
        Reason: "Customer request",
    })
    
    assert.Equal(t, "completed", refundResp.Status)
    assert.Equal(t, int64(5000), refundResp.Amount)
}
```

---

## Test Database Setup

### Docker Test Database

```bash
# Start test database
docker run -d \
  --name pay2go-test-db \
  -e POSTGRES_USER=test_user \
  -e POSTGRES_PASSWORD=test_pass \
  -e POSTGRES_DB=pay2go_test \
  -p 5434:5432 \
  postgres:16-alpine

# Run migrations
migrate -path ./migrations \
  -database "postgres://test_user:test_pass@localhost:5434/pay2go_test?sslmode=disable" \
  up
```

### Test Database Helpers

```go
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()
    
    dsn := "postgres://test_user:test_pass@localhost:5434/pay2go_test?sslmode=disable"
    db, err := sql.Open("postgres", dsn)
    require.NoError(t, err)
    
    // Run migrations
    runMigrations(t, db)
    
    return db
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
    t.Helper()
    
    // Clean all tables
    tables := []string{"refunds", "transactions", "partners"}
    for _, table := range tables {
        _, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
        require.NoError(t, err)
    }
}
```

---

## Mocking

### Repository Mocks

```go
type MockTransactionRepository struct {
    mock.Mock
}

func (m *MockTransactionRepository) Create(ctx context.Context, tx *entities.Transaction) error {
    args := m.Called(ctx, tx)
    return args.Error(0)
}

func (m *MockTransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Transaction, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*entities.Transaction), args.Error(1)
}
```

### Payment Gateway Mocks

```go
type MockPaymentGateway struct {
    mock.Mock
}

func (m *MockPaymentGateway) ProcessPayment(ctx context.Context, tx *entities.Transaction) (string, error) {
    args := m.Called(ctx, tx)
    return args.String(0), args.Error(1)
}
```

---

## Test Data Builders

```go
// Test data builder pattern
type TransactionBuilder struct {
    transaction *entities.Transaction
}

func NewTransactionBuilder() *TransactionBuilder {
    amount, _ := valueobjects.NewMoney(10000, "USD")
    tx, _ := entities.NewTransaction(
        uuid.New(),
        amount,
        valueobjects.PaymentMethodCreditCard,
        valueobjects.PaymentProviderStripe,
        "Test",
        uuid.New().String(),
        nil,
    )
    
    return &TransactionBuilder{transaction: tx}
}

func (b *TransactionBuilder) WithAmount(amount int64) *TransactionBuilder {
    money, _ := valueobjects.NewMoney(amount, b.transaction.Amount.Currency)
    b.transaction.Amount = money
    return b
}

func (b *TransactionBuilder) WithStatus(status string) *TransactionBuilder {
    b.transaction.Status = status
    return b
}

func (b *TransactionBuilder) Build() *entities.Transaction {
    return b.transaction
}

// Usage
tx := NewTransactionBuilder().
    WithAmount(50000).
    WithStatus("completed").
    Build()
```

---

## Performance Testing

### Load Testing with k6

```javascript
// load_test.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    stages: [
        { duration: '30s', target: 20 },  // Ramp up
        { duration: '1m', target: 100 },  // Stay at 100 users
        { duration: '30s', target: 0 },   // Ramp down
    ],
};

export default function() {
    const url = 'http://localhost:8080/api/v1/transactions';
    const payload = JSON.stringify({
        amount: 10000,
        currency: 'USD',
        payment_method: 'credit_card',
        payment_provider: 'stripe',
        description: 'Load test',
        idempotency_key: `load-${__VU}-${__ITER}`,
    });
    
    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer test-api-key',
        },
    };
    
    let res = http.post(url, payload, params);
    
    check(res, {
        'status is 201': (r) => r.status === 201,
        'response time < 500ms': (r) => r.timings.duration < 500,
    });
    
    sleep(1);
}
```

Run load test:
```bash
k6 run load_test.js
```

---

## Test Coverage Goals

### Minimum Coverage Targets

- **Overall**: 80%
- **Domain Layer**: 90%+ (critical business logic)
- **Use Cases**: 85%+ (orchestration logic)
- **Handlers**: 75%+ (HTTP layer)
- **Repositories**: 70%+ (data access)

### Measuring Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage by package
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

---

## Continuous Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: test_user
          POSTGRES_PASSWORD: test_pass
          POSTGRES_DB: pay2go_test
        ports:
          - 5434:5432
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: make test-coverage
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

---

## Best Practices

### Test Organization

✅ **DO**:
- Use table-driven tests for multiple scenarios
- Follow AAA pattern (Arrange, Act, Assert)
- Use descriptive test names (`TestEntity_Method_Scenario`)
- Create test helpers for common setup
- Use builders for complex test data
- Mock external dependencies
- Clean up resources (defer cleanup)

❌ **DON'T**:
- Test implementation details
- Create interdependent tests
- Use sleep() for timing
- Hardcode test data
- Share state between tests
- Skip error checking in tests

### Test Naming Convention

```go
// Format: Test<Entity>_<Method>_<Scenario>
func TestTransaction_MarkAsCompleted_Success(t *testing.T) {}
func TestTransaction_MarkAsCompleted_AlreadyCompleted(t *testing.T) {}
func TestMoney_Add_DifferentCurrencies(t *testing.T) {}
```

---

## Troubleshooting Tests

### Common Issues

**Test Database Connection Failed**:
```bash
# Check if test database is running
docker ps | grep pay2go-test-db

# Restart test database
docker restart pay2go-test-db
```

**Flaky Tests**:
```bash
# Run specific test multiple times
go test -run TestName -count=10 ./...
```

**Race Conditions**:
```bash
# Run tests with race detector
go test -race ./...
```

---

## Next Steps

1. Implement remaining unit tests for all domain entities
2. Add integration tests for all use cases
3. Create E2E test suite for critical flows
4. Set up CI/CD pipeline with automated testing
5. Configure code coverage reporting
6. Add performance benchmarks

---

For questions about testing, contact the QA team.
