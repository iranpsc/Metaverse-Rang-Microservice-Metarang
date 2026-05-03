# Social Service — Testing

## Unit tests (no database)

Runs handler and service tests with in-memory gRPC (`internal/testutil/grpc_bufconn.go`) and mocks (`internal/testutil/mocks.go`):

```bash
cd services/social-service
go test ./internal/handler/... ./internal/service/... -race -count=1
```

## Coverage gate (≥70%)

Combined **handler + service** statement coverage is enforced by the repo Makefile:

```bash
make test-coverage-social
```

This runs:

```bash
go test ./internal/handler/... ./internal/service/... -race -coverprofile=coverage.out -covermode=atomic
```

## Repository integration tests (optional)

Repository tests call a real MySQL when `TEST_MYSQL_DSN` is set; otherwise they skip.

Example:

```bash
export TEST_MYSQL_DSN='metargb_user:metargb_password@tcp(127.0.0.1:3306)/metargb_db?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci'
go test ./internal/repository/... -count=1
```

## Full internal suite

```bash
go test ./internal/... -race -count=1
```
