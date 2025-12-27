# API Documentation

## Pay2Go Payment Orchestration API

Version: 1.0  
Base URL: `http://localhost:8080/api/v1`

### Authentication

All protected endpoints require an API key passed in the `Authorization` header:

```
Authorization: Bearer <your-api-key>
```

API keys are issued per partner and can be managed through the partner management interface.

### Rate Limiting

- **Rate Limit**: 100 requests per minute per partner
- **Headers**: 
  - `X-RateLimit-Limit`: Maximum requests allowed
  - `X-RateLimit-Remaining`: Remaining requests
  - `X-RateLimit-Reset`: Time when limit resets (Unix timestamp)

### Error Responses

All errors follow this format:

```json
{
  "error": "Error message description"
}
```

Common HTTP status codes:
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (missing or invalid API key)
- `404` - Not Found (resource doesn't exist)
- `429` - Too Many Requests (rate limit exceeded)
- `500` - Internal Server Error

---

## Endpoints

### Health Checks

#### GET /api/v1/health
Check if the service is running.

**Response**: `200 OK`
```json
{
  "status": "ok"
}
```

#### GET /api/v1/health/ready
Check if the service is ready to accept requests.

**Response**: `200 OK`
```json
{
  "status": "ready",
  "database": "connected"
}
```

#### GET /api/v1/health/live
Kubernetes liveness probe endpoint.

**Response**: `200 OK`
```json
{
  "status": "alive"
}
```

---

### Transactions

#### POST /api/v1/transactions
Create a new payment transaction.

**Headers**:
- `Authorization: Bearer <api-key>` (required)
- `Content-Type: application/json`

**Request Body**:
```json
{
  "amount": 10000,
  "currency": "USD",
  "payment_method": "credit_card",
  "payment_provider": "stripe",
  "description": "Order #12345",
  "idempotency_key": "unique-key-123",
  "metadata": {
    "order_id": "12345",
    "customer_email": "customer@example.com"
  }
}
```

**Fields**:
- `amount` (int64, required): Amount in cents (e.g., 10000 = $100.00)
- `currency` (string, required): ISO 4217 currency code (USD, EUR, GBP)
- `payment_method` (string, required): Payment method (`credit_card`, `debit_card`, `bank_transfer`, `digital_wallet`)
- `payment_provider` (string, required): Payment provider (`stripe`, `paypal`, `adyen`, `manual`)
- `description` (string, required): Transaction description
- `idempotency_key` (string, required): Unique key to prevent duplicate transactions
- `metadata` (object, optional): Additional metadata as key-value pairs

**Response**: `201 Created`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "partner_id": "partner-uuid",
  "amount": 10000,
  "currency": "USD",
  "status": "pending",
  "payment_method": "credit_card",
  "payment_provider": "stripe",
  "description": "Order #12345",
  "idempotency_key": "unique-key-123",
  "metadata": {
    "order_id": "12345",
    "customer_email": "customer@example.com"
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

---

#### GET /api/v1/transactions/:id
Get a specific transaction by ID.

**Headers**:
- `Authorization: Bearer <api-key>` (required)

**Path Parameters**:
- `id` (UUID, required): Transaction ID

**Response**: `200 OK`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "partner_id": "partner-uuid",
  "amount": 10000,
  "currency": "USD",
  "status": "completed",
  "payment_method": "credit_card",
  "payment_provider": "stripe",
  "provider_transaction_id": "stripe_ch_3abc123",
  "description": "Order #12345",
  "idempotency_key": "unique-key-123",
  "metadata": {
    "order_id": "12345"
  },
  "completed_at": "2024-01-15T10:31:00Z",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:31:00Z"
}
```

---

#### GET /api/v1/transactions
List all transactions for the authenticated partner.

**Headers**:
- `Authorization: Bearer <api-key>` (required)

**Query Parameters**:
- `status` (string, optional): Filter by status (`pending`, `completed`, `failed`, `refunded`)
- `from_date` (string, optional): Filter by creation date (RFC3339 format)
- `to_date` (string, optional): Filter by creation date (RFC3339 format)
- `limit` (int, optional): Number of results per page (default: 20, max: 100)
- `offset` (int, optional): Pagination offset (default: 0)

**Example Request**:
```
GET /api/v1/transactions?status=completed&limit=10
```

**Response**: `200 OK`
```json
{
  "transactions": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "amount": 10000,
      "currency": "USD",
      "status": "completed",
      "payment_method": "credit_card",
      "description": "Order #12345",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

---

#### POST /api/v1/transactions/:id/process
Process a pending transaction through the payment gateway.

**Headers**:
- `Authorization: Bearer <api-key>` (required)

**Path Parameters**:
- `id` (UUID, required): Transaction ID

**Response**: `200 OK`
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "status": "completed",
  "provider_transaction_id": "stripe_ch_3abc123",
  "completed_at": "2024-01-15T10:31:00Z"
}
```

**Error Response**: `400 Bad Request`
```json
{
  "error": "Transaction is not in pending status"
}
```

---

#### POST /api/v1/transactions/:id/refund
Refund a completed transaction (fully or partially).

**Headers**:
- `Authorization: Bearer <api-key>` (required)
- `Content-Type: application/json`

**Path Parameters**:
- `id` (UUID, required): Transaction ID

**Request Body**:
```json
{
  "amount": 5000,
  "reason": "Customer requested refund"
}
```

**Fields**:
- `amount` (int64, required): Refund amount in cents (must not exceed transaction amount)
- `reason` (string, required): Reason for the refund

**Response**: `200 OK`
```json
{
  "refund_id": "refund-uuid",
  "transaction_id": "123e4567-e89b-12d3-a456-426614174000",
  "amount": 5000,
  "currency": "USD",
  "status": "completed",
  "reason": "Customer requested refund",
  "provider_refund_id": "stripe_re_3xyz789",
  "created_at": "2024-01-15T11:00:00Z"
}
```

**Business Rules**:
- Transaction must be in `completed` status
- Refund must be within 90 days of transaction completion
- Total refunds cannot exceed original transaction amount
- Partial refunds are allowed

---

## Payment Methods

Supported payment methods:
- `credit_card` - Credit Card
- `debit_card` - Debit Card
- `bank_transfer` - Bank Transfer
- `digital_wallet` - Digital Wallet (Apple Pay, Google Pay, etc.)

## Payment Providers

Supported payment providers:
- `stripe` - Stripe
- `paypal` - PayPal
- `adyen` - Adyen
- `manual` - Manual Processing

## Transaction States

```
pending → processing → completed
        ↓              ↓
      failed      refunded
```

- `pending`: Transaction created, awaiting processing
- `processing`: Payment is being processed by provider
- `completed`: Payment successfully processed
- `failed`: Payment processing failed
- `refunded`: Transaction has been refunded

---

## Examples

### Create and Process a Transaction

```bash
# 1. Create transaction
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 10000,
    "currency": "USD",
    "payment_method": "credit_card",
    "payment_provider": "stripe",
    "description": "Order #12345",
    "idempotency_key": "unique-key-'$(date +%s)'"
  }'

# 2. Process transaction
curl -X POST http://localhost:8080/api/v1/transactions/{transaction-id}/process \
  -H "Authorization: Bearer your-api-key"

# 3. Get transaction status
curl -X GET http://localhost:8080/api/v1/transactions/{transaction-id} \
  -H "Authorization: Bearer your-api-key"
```

### Refund a Transaction

```bash
curl -X POST http://localhost:8080/api/v1/transactions/{transaction-id}/refund \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 5000,
    "reason": "Customer requested refund"
  }'
```

### List Transactions

```bash
# Get all completed transactions
curl -X GET "http://localhost:8080/api/v1/transactions?status=completed&limit=10" \
  -H "Authorization: Bearer your-api-key"
```

---

## Webhooks (Future Implementation)

Webhooks will be available to notify partners of transaction state changes:

- `transaction.completed` - Transaction successfully completed
- `transaction.failed` - Transaction processing failed
- `transaction.refunded` - Transaction refunded

---

## Best Practices

1. **Idempotency**: Always include a unique `idempotency_key` to prevent duplicate transactions
2. **Error Handling**: Implement retry logic with exponential backoff for 5xx errors
3. **Security**: Never expose API keys in client-side code
4. **Monitoring**: Monitor rate limit headers to avoid hitting limits
5. **Testing**: Use test mode API keys in development environments

---

## Support

For API support, contact: support@pay2go.com
