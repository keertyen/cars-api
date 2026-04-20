# cars-api

REST API for managing car listings — CRUD over a collection of cars backed by Firestore, with an in-memory cache layer and cursor-based pagination.

## Stack
- **Go 1.22** (enforced via `go` directive in `go.mod` — no `toolchain` pin)
- **Fiber v2** — HTTP framework
- **Firestore** — persistence (emulated locally via Firebase CLI)
- **go-playground/validator** — struct validation
- **slog** — structured JSON logging
- **testify** — assertions in tests

## API endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Liveness probe |
| `GET` | `/ready` | Readiness probe |
| `POST` | `/v1/cars` | Create a car |
| `GET` | `/v1/cars` | List cars (cursor pagination: `?pageSize=&pageToken=`) |
| `GET` | `/v1/cars/:id` | Get a car by ID |
| `PUT` | `/v1/cars/:id` | Update a car (partial — omitted fields are left unchanged) |
| `DELETE` | `/v1/cars/:id` | Delete a car |

## Project layout

```
cmd/server/main.go              entry point, wiring, graceful shutdown
config/config.go                env-var config (no external lib)
internal/model/car.go           domain types + request/response structs
internal/cache/cache.go         in-process TTL cache (sync.RWMutex, no deps)
internal/store/store.go         Store interface + ErrNotFound sentinel
internal/store/firestore/store.go  Firestore implementation
internal/service/service.go     Service interface
internal/service/car.go         business logic + cache orchestration
internal/service/car_test.go    unit tests (mock store)
internal/api/handler.go         HTTP handlers + validation
internal/api/router.go          Fiber app + middleware wiring
internal/api/response.go        ErrorResponse type + traceID helper
internal/api/handler_test.go    handler tests (mock service)
```

## Architecture

Request → Fiber middleware (RequestID, Recover, Logger, RateLimiter, Compress) → Handler (validate) → Service (cache → store) → Firestore

- `Service` and `Store` are interfaces. Tests mock them directly — no real DB needed.
- Cache keys: `car:<id>` (default TTL from `CACHE_TTL_SECONDS`) and `cars:list:<size>:<token>` (30s TTL). All list keys bust together on any write via `DeletePrefix`.
- `UpdateCarRequest` uses pointer fields — `nil` means "not provided", so PUT behaves like PATCH.
- Firestore pagination uses `StartAfter(docSnapshot)` — avoids the O(n) read cost of offset pagination.
- Per-op context deadlines: 5s for single-doc ops, 10s for list.

## Docker

Two-stage build: `golang:1.22-alpine` compiles a static binary (`CGO_ENABLED=0`), then copied into a minimal `alpine:3.19` runtime image. Exposes port `8080`.

Preferred: use docker compose — it starts the Firestore emulator as a dependency and wires `FIRESTORE_EMULATOR_HOST` automatically:

```bash
make docker-up    # docker compose up --build
make docker-down  # docker compose down --volumes
```

To run the image standalone (bring your own Firestore):

```bash
docker build -t cars-api .
docker run -p 8080:8080 \
  -e GCP_PROJECT_ID=my-project \
  -e FIRESTORE_EMULATOR_HOST=host.docker.internal:8081 \
  cars-api
```

## Running locally

Requires **Firebase CLI** (`npm install -g firebase-tools`). `make run` shells out to `firebase emulators:start` — version quirks with the Firestore emulator are common; Firebase CLI ≥ 13.x is recommended.

```bash
make run          # starts Firestore emulator + server on :8080
make test         # unit tests with race detector
make test-cover   # generates coverage.out + opens coverage.html (~75% api, ~80% service; no CI threshold)
make lint         # golangci-lint (see .golangci.yml — govet, errcheck, staticcheck)
```

## Env vars

See [.env.example](.env.example) for a ready-to-copy template.

| Var | Default |
|-----|---------|
| `PORT` | `8080` |
| `GCP_PROJECT_ID` | `cars-api-local` |
| `FIRESTORE_EMULATOR_HOST` | — |
| `RATE_LIMIT_RPS` | `10` |
| `RATE_LIMIT_BURST` | `20` |
| `CACHE_TTL_SECONDS` | `300` |

## Conventions

- Comments explain *why*, not *what* — if it's obvious from the code, no comment.
- No godoc stubs on unexported or obvious methods.
- Errors always wrapped with `fmt.Errorf("layer.Op: %w", err)` for traceable chains.
- `store.ErrNotFound` is the canonical sentinel — handlers check with `errors.Is`.
- Tests are fully isolated: service tests use `mockStore`, handler tests use `mockService`.
