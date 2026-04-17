# cars-api

REST API for a car inventory. Go + Firestore (emulated locally).

## Setup

```bash
make run                  # starts emulator + server
docker-compose up --build # full docker alternative
```

## Endpoints

```
GET    /health
POST   /v1/cars
GET    /v1/cars?page_size=20&page_token=<cursor>
GET    /v1/cars/:id
PUT    /v1/cars/:id
DELETE /v1/cars/:id
```

PUT behaves like PATCH — only fields you send are updated.

## Examples

```bash
curl -X POST http://localhost:8080/v1/cars \
  -H "Content-Type: application/json" \
  -d '{"make":"Toyota","model":"Camry","year":2022,"color":"Blue","price":28500,"status":"available"}'

curl http://localhost:8080/v1/cars?page_size=5
curl http://localhost:8080/v1/cars/<id>
curl -X PUT http://localhost:8080/v1/cars/<id> -H "Content-Type: application/json" -d '{"status":"sold"}'
curl -X DELETE http://localhost:8080/v1/cars/<id>
```

## Env vars

| Var | Default |
|-----|---------|
| `PORT` | `8080` |
| `GCP_PROJECT_ID` | `cars-api-local` |
| `FIRESTORE_EMULATOR_HOST` | — |
| `RATE_LIMIT_RPS` | `10` |
| `RATE_LIMIT_BURST` | `20` |
| `CACHE_TTL_SECONDS` | `300` |

## Tests

```bash
make test
```

## Notes

List pagination uses Firestore `StartAfter` (cursor) rather than offset — avoids the O(n) read cost Firestore charges when skipping documents.

Cache is in-process — not shared across replicas. Swap for Redis in `service.New` if needed.
