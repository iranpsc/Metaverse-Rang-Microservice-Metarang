# Features Service

The Features Service is a Go microservice that handles all feature-related operations in the MetaRGB platform, including marketplace transactions, buy/sell requests, hourly profits, and map data.

## Overview

This service provides gRPC endpoints for:
- **Feature Marketplace**: Buy/sell features, manage buy/sell requests
- **Feature Management**: List, view, and manage user-owned features
- **Hourly Profits**: Withdraw and manage feature hourly profits
- **Maps**: Retrieve map polygons and feature rollups
- **Buildings**: Manage feature buildings and 3D environments

## Architecture

The service follows a layered architecture:

```
Handler Layer (internal/handler/)
  ↓
Service Layer (internal/service/)
  ↓
Repository Layer (internal/repository/)
  ↓
Database (MySQL)
```

### Key Components

- **Handlers**: gRPC request/response conversion, validation, error mapping
- **Services**: Business logic orchestration, cross-service communication
- **Repositories**: Data access layer with direct SQL queries
- **Clients**: gRPC clients for Commercial Service and Notification Service
- **Events**: Redis-based event broadcasting for real-time updates

## Dependencies

### External Services

- **Commercial Service** (gRPC): Wallet operations, transactions
- **Notification Service** (gRPC): Send notifications to users
- **Auth Service** (gRPC): Token validation
- **3D Meta API** (HTTP): Building generation

### Infrastructure

- **MySQL**: Primary database for features, requests, trades
- **Redis**: Event broadcasting (Pub/Sub)

## Configuration

Copy `config.env.sample` to `config.env` and configure:

```env
# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_DATABASE=metargb_db

# gRPC Server
PORT=50051

# Metrics
METRICS_PORT=9090

# External Services
COMMERCIAL_SERVICE_ADDR=commercial-service:50052
NOTIFICATIONS_SERVICE_ADDR=notifications-service:50058
AUTH_SERVICE_ADDR=auth-service:50051

# Redis (for event broadcasting)
REDIS_ADDR=redis:6379
REDIS_PASSWORD=
BROADCAST_CHANNEL=feature-events

# Commercial Service Configuration
COMMERCIAL_SERVICE_TIMEOUT=3s
COMMERCIAL_SERVICE_RETRIES=3

# 3D Meta API
THREE_D_META_URL=http://3d-meta-api
```

## Running the Service

### Local Development

```bash
# Load environment variables
source config.env

# Run the service
go run cmd/server/main.go
```

### Docker

```bash
docker build -f Dockerfile.dev -t features-service:dev .
docker run --env-file config.env features-service:dev
```

## API Documentation

### gRPC Services

The service exposes the following gRPC services:

- `FeatureService`: Feature browsing and details
- `FeatureMarketplaceService`: Buy/sell operations, requests
- `FeatureProfitService`: Hourly profit management
- `MapsService`: Map data retrieval
- `BuildingService`: Building management

See `api-docs/features-service/` for detailed REST API documentation (via gRPC Gateway).

## Metrics

Prometheus metrics are exposed on `/metrics` endpoint (port 9090 by default):

- `metargb_features_buy_requests_total`: Buy requests by status (accepted/rejected/cancelled)
- `metargb_features_sell_requests_total`: Total sell requests
- `metargb_features_trades_total`: Trades by type (limited/rgb/user)
- `metargb_features_trade_value_psc`: Trade values in PSC (histogram)
- `metargb_features_trade_value_irr`: Trade values in IRR (histogram)
- `metargb_features_buy_request_locked_assets_psc`: Locked PSC assets (gauge)
- `metargb_features_buy_request_locked_assets_irr`: Locked IRR assets (gauge)

## Event Broadcasting

The service broadcasts `FeatureStatusChanged` events via Redis when:
- A sell request is created
- A sell request is deleted
- A feature is purchased (all three paths: limited, RGB, user-to-user)
- A buy request is accepted

Events are published to the channel specified by `BROADCAST_CHANNEL` (default: `feature-events`).

## Notifications

The service sends notifications for:
- Buy request creation (to buyer and seller)
- Buy request acceptance (to buyer and seller)
- Sell request creation (to seller)
- Feature purchase completion (to buyer)
- Hourly profit withdrawal (to user)

## Error Handling

Errors are mapped to gRPC status codes:
- `InvalidArgument`: Validation errors, bad input
- `NotFound`: Resource not found
- `Unauthenticated`: Missing/invalid token
- `PermissionDenied`: Authorization failure
- `FailedPrecondition`: Business rule violation
- `Internal`: Unexpected server errors

## Testing

See `TESTING.md` for testing guidelines and examples.

## Development

### Project Structure

```
features-service/
├── cmd/server/          # Application entry point
├── internal/
│   ├── handler/         # gRPC handlers
│   ├── service/         # Business logic
│   ├── repository/      # Data access
│   ├── client/          # External service clients
│   ├── events/          # Event broadcasting
│   ├── metrics/         # Prometheus metrics
│   ├── models/          # Domain models
│   └── constants/       # Constants and configuration
├── pkg/                 # Public packages
├── config.env.sample    # Configuration template
└── Dockerfile           # Container definition
```

### Adding New Features

1. Define proto messages in `shared/proto/features/`
2. Run `make proto` to generate Go code
3. Implement repository methods (data access)
4. Implement service methods (business logic)
5. Implement handler methods (gRPC interface)
6. Register handler in `main.go`
7. Add tests

## License

Part of the MetaRGB microservices platform.

