# PocketBase (fork with distributed realtime and multi-database support)

This repository is a community-maintained fork of `pocketbase/pocketbase` that adds:

- Distributed realtime subscriptions via Redis Pub/Sub (multi-node ready)
- Multi-database support: MySQL 8+, PostgreSQL 13+, and SQLite (default)
- Ready-to-use examples and docker-compose for local development

All core PocketBase features remain available: embedded database (SQLite), REST API, admin UI, file and user management, JS VM extensions, etc.

For upstream docs, see https://pocketbase.io/docs.

## What's included in this fork

- Added
  - Distributed realtime: cross-instance broadcasting through Redis Pub/Sub
  - Multi-database support beyond SQLite (MySQL, PostgreSQL)
  - Environment variables and docker-compose samples for quick setup
- Unchanged (compatibility goals)
  - Default single-node behavior and upstream APIs
  - Backward-compatible usage if you don’t enable Redis or external DBs

## Quick Start

### A) Single node (SQLite, no dependencies)

```bash
# Go 1.23+
go mod tidy

# Run example app (same style as upstream)
cd examples/base
GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) CGO_ENABLED=0 go build
./base serve
```

Access:
- Admin/API: http://localhost:8090
- Health: http://localhost:8090/api/health

### B) Multi-database + Redis (docker-compose)

```bash
# Start MySQL, PostgreSQL and Redis services
docker-compose up -d

# Prepare environment variables
cp env.example .env
# Edit .env to select DB and (optionally) enable Redis-based features

# Example: MySQL
export PB_DB_TYPE=mysql
export PB_DB_DSN="root:rootpassword@tcp(localhost:3306)/pocketbase?parseTime=true&charset=utf8mb4&loc=Local"

# Start the app (example entrypoint)
go run examples/base/main.go serve
```

## Configuration

You can configure the app via environment variables (see `env.example`):

- Database
  - `PB_DB_TYPE`: `sqlite` | `mysql` | `postgres`
  - `PB_DB_DSN`:
    - MySQL: `user:pass@tcp(host:port)/db?parseTime=true&charset=utf8mb4&loc=Local`
    - PostgreSQL: `postgres://user:pass@host:port/db?sslmode=disable&search_path=public`
    - SQLite: usually not required (defaults used)
  - Pool (optional): `PB_DB_MAX_OPEN`, `PB_DB_MAX_IDLE`, `PB_DB_CONN_MAX_LIFETIME`
- Directories (optional): `PB_DATA_DIR`, `PB_PUBLIC_DIR`
- Distributed features (optional):
  - `PB_REDIS_URL`: e.g. `redis://localhost:6379/0` or `redis://:password@localhost:6379/0`
  - `PB_REALTIME_DISTRIBUTED=true` to enable distributed realtime
  - `PB_RATELIMIT_DISTRIBUTED=true` if you enable distributed rate limiting

## Distributed Realtime (Redis)

Goal: make SSE-based realtime events work seamlessly across multiple app instances. Any event produced on one instance is propagated via Redis Pub/Sub and delivered to subscribers connected to other instances.

Recommended topology: N application instances + 1 Redis (or Redis cluster).

Steps:
1) Start Redis (docker-compose provides a service).
2) Set env vars:
   ```bash
   export PB_REDIS_URL=redis://localhost:6379/0
   export PB_REALTIME_DISTRIBUTED=true
   ```
3) Run multiple instances and verify that updates on one node are received by subscribers connected to other nodes.

Notes:
- Redis Pub/Sub is best-effort and non-durable. If you need guaranteed delivery/persistence, consider an additional queue or event log.
- For production, use Redis Sentinel/Cluster for HA, and secure your Redis network access and credentials.

## Multi-Database Support

Choose your database by setting `PB_DB_TYPE` and `PB_DB_DSN`.

- SQLite (default, zero deps)
- MySQL 8+:
  ```bash
  export PB_DB_TYPE=mysql
  export PB_DB_DSN="root:rootpassword@tcp(localhost:3306)/pocketbase?parseTime=true&charset=utf8mb4&loc=Local"
  ```
- PostgreSQL 13+:
  ```bash
  export PB_DB_TYPE=postgres
  export PB_DB_DSN="postgres://pocketbase:pocketbase123@localhost:5432/pocketbase?sslmode=disable&search_path=public"
  ```

Recommendations:
- Tune pool settings: `PB_DB_MAX_OPEN`, `PB_DB_MAX_IDLE`, `PB_DB_CONN_MAX_LIFETIME`.
- Validate connectivity with the health endpoint: `curl http://localhost:8090/api/health`.

## Testing

```bash
# Run all tests
go test ./...

# Optionally target API/realtime packages
go test ./apis -run Realtime -v
```

If using docker-compose services, verify they are up:
```bash
docker-compose ps
```

## Migration & Compatibility

- Existing PocketBase users can adopt this fork without changes in single-node mode.
- To enable distributed realtime, add Redis and set `PB_REDIS_URL` + `PB_REALTIME_DISTRIBUTED=true`.
- To switch databases, set `PB_DB_TYPE`/`PB_DB_DSN` and migrate/import your data.
- API behavior aims to remain compatible with upstream (unless otherwise documented).

## Build Targets (SQLite driver)

When building statically with the pure Go SQLite driver, supported targets include (subject to upstream driver support):

```
darwin  amd64
darwin  arm64
freebsd amd64
freebsd arm64
linux   386
linux   amd64
linux   arm
linux   arm64
linux   loong64
linux   ppc64le
linux   riscv64
linux   s390x
windows 386
windows amd64
windows arm64
```

## Security

If you discover a security vulnerability, please open a private report or contact the maintainers. We will address issues promptly and credit reporters in release notes.

## License

This fork and the upstream PocketBase are licensed under the MIT License (see `LICENSE.md`).

## Credits

- Upstream: https://github.com/pocketbase/pocketbase
- Thanks to all contributors of the original project.
