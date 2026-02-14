# GoBilling - Subscription Billing System

A production-ready subscription billing system built with Go, PostgreSQL, and Redis.

## Architecture

- **Modular Monolith** - Clean domain boundaries with potential microservices migration path
- **Event-Driven** - Transactional outbox pattern for reliable event delivery
- **ACID Compliant** - All financial operations within database transactions
- **Horizontally Scalable** - Stateless application layer with connection pooling

## Tech Stack

- **Go 1.22+** - Application runtime
- **PostgreSQL 15+** - Primary database with JSONB support
- **Redis 7+** - Caching and idempotency keys
- **Chi Router** - HTTP routing and middleware
- **pgx/v5** - High-performance PostgreSQL driver
- **OpenTelemetry** - Distributed tracing
- **Prometheus** - Metrics collection

## Project Structure

```
gobilling/
├── cmd/
│   └── server/          # Application entrypoint
├── internal/
│   ├── customer/        # Customer domain
│   ├── product/         # Product & Plan domain
│   ├── subscription/    # Subscription lifecycle
│   ├── invoice/         # Invoice generation
│   ├── payment/         # Payment processing
│   ├── ledger/          # Transaction ledger
│   ├── event/           # Domain events
│   ├── webhook/         # Webhook delivery
│   ├── platform/        # Infrastructure
│   │   ├── config/
│   │   ├── database/
│   │   ├── cache/
│   │   ├── http/
│   │   └── errors/
│   └── pkg/             # Shared utilities
│       ├── id/
│       ├── money/
│       ├── clock/
│       └── pagination/
├── migrations/          # Database migrations
├── docs/               # Documentation
└── docker-compose.yml  # Local development
```

## Getting Started

### Prerequisites

- Go 1.22+
- PostgreSQL 15+
- Redis 7+
- golang-migrate CLI

### Installation

1. Clone the repository
2. Copy `.env.example` to `.env` and configure
3. Run database migrations:
   ```bash
   make migrate-up
   ```
4. Start the application:
   ```bash
   make run
   ```

### Docker Development

```bash
docker-compose up -d
make migrate-up
make run
```

## API Documentation

Base URL: `http://localhost:8080/v1`

### Authentication

All requests require API key authentication:

```
Authorization: Bearer sk_live_...
```

### Core Endpoints

- `POST /v1/customers` - Create customer
- `GET /v1/customers` - List customers
- `POST /v1/products` - Create product
- `POST /v1/plans` - Create pricing plan
- `POST /v1/subscriptions` - Create subscription
- `GET /v1/invoices` - List invoices
- `POST /v1/payments/{id}/refund` - Refund payment

See `docs/api/api-spec.md` for full API documentation.

## Domain Models

### Customer
- Lifecycle: ACTIVE → SUSPENDED → DELETED
- Soft delete only
- Email uniqueness enforced

### Subscription
- States: TRIALING → ACTIVE → PAST_DUE → CANCELED
- Automatic renewal via background worker
- Payment retry with exponential backoff

### Invoice
- States: DRAFT → OPEN → PAID/VOID/UNCOLLECTIBLE
- Sequential numbering per tenant
- Immutable after finalization

### Payment
- States: PENDING → SUCCEEDED/FAILED
- Retry schedule: 1h, 6h, 24h, 72h
- Full refund support

## Background Workers

- **RenewalWorker** - Processes subscription renewals (1 min interval)
- **PaymentRetryWorker** - Retries failed payments (1 min interval)
- **OutboxWorker** - Dispatches domain events (500ms interval)
- **WebhookDeliveryWorker** - Delivers webhooks (1 sec interval)

## Security

- API key authentication with SHA-256 hashing
- RBAC authorization with resource scoping
- Rate limiting per API key
- Input validation and sanitization
- SQL injection prevention via parameterized queries
- TLS 1.2+ for all connections

## Observability

- **Logging** - Structured JSON logs with slog
- **Metrics** - Prometheus metrics on `/metrics`
- **Tracing** - OpenTelemetry with Jaeger export
- **Health Checks** - `/health` (liveness), `/ready` (readiness)

## Testing

```bash
make test
```

## Deployment

See `docs/deployment/` for production deployment guides.

## License

Proprietary
