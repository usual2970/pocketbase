# Development Setup

This document describes how to set up a development environment for testing multi-database support.

## Quick Start

1. **Start Database Services**
   ```bash
   docker-compose up -d
   ```

2. **Configure Environment**
   ```bash
   cp env.example .env
   # Edit .env to select your preferred database
   ```

3. **Test MySQL Connection**
   ```bash
   export PB_DB_TYPE=mysql
   export PB_DB_DSN="root:rootpassword@tcp(localhost:3306)/pocketbase?parseTime=true&charset=utf8mb4&loc=Local"
   go run main.go serve
   ```

4. **Test PostgreSQL Connection**
   ```bash
   export PB_DB_TYPE=postgres
   export PB_DB_DSN="postgres://pocketbase:pocketbase123@localhost:5432/pocketbase?sslmode=disable&search_path=public"
   go run main.go serve
   ```

## Database Services

### MySQL 8.0
- **Host**: localhost:3306
- **Database**: pocketbase
- **Username**: pocketbase / root
- **Password**: pocketbase123 / rootpassword

### PostgreSQL 13
- **Host**: localhost:5432
- **Database**: pocketbase
- **Username**: pocketbase
- **Password**: pocketbase123

## Health Check

Once the services are running, you can verify database connectivity:

```bash
curl http://localhost:8090/api/health
```

For superuser access, authenticate first and check the `dbConnected` field in the response.

## Cleanup

```bash
# Stop and remove containers
docker-compose down

# Remove volumes (WARNING: deletes all data)
docker-compose down -v
```

## Troubleshooting

- **Connection refused**: Ensure services are running with `docker-compose ps`
- **Authentication failed**: Check username/password in the DSN
- **Database not found**: Services create databases automatically on first run
- **Version errors**: Ensure you're using MySQL 8.0+ and PostgreSQL 13+
