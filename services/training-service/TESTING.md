# Training service tests

```bash
go test ./internal/service ./internal/handler ./internal/repository -count=1 -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out
```

Primary targets (unit tests with mocks):

- `internal/service` — business logic (mock repositories)
- `internal/handler` — gRPC handlers via in-memory bufconn + real services + mocks
- `internal/repository` — SQL layer smoke tests using `go-sqlmock` (subset of queries)

Integration tests against MySQL can be added later with a build tag (e.g. `integration`).
