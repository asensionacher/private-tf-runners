# Agents

## Build Commands

### Backend (Go)
```bash
# Server (CGO_ENABLED=1 required for SQLite)
go build -o server ./cmd/server

# Runner (CGO_ENABLED=0 for minimal image)
CGO_ENABLED=0 go build -ldflags="-w -s" -o runner ./cmd/runner
```

### Docker
```bash
# Backend image
docker build -t private-tf-runners-backend -f ./Dockerfile .

# Runner image (no cache for fresh build)
docker build -t private-tf-runners-runner --no-cache -f ./cmd/runner/Dockerfile .
```

### Frontend (React)
```bash
cd frontend
npm run dev      # Development server
npm run build    # Production build (tsc -b && vite build)
npm run lint     # ESLint
```

## User Roles and Permissions

Three roles defined in `internal/models/models.go`:
- `viewer` - Read-only (stacks, runs, runners)
- `operator` - Create/delete stacks, runs, runners, reset tokens
- `admin` - Full access including user management

Frontend permission utilities at `frontend/src/lib/permissions.ts`. Use `useAuth().can(permission)` to check.

## Backend Schemas

Terraform backend schemas in `internal/backends/config.go`. Supported: `azurerm`, `s3`, `http`, `pg`, `kubernetes`, `gcs`.

Use `GetBackendSchema(name)` and `ValidateConfig(name, json)`.

## Database Schema

Schema in `internal/database/database.go` in `migrate()` function. When modifying CREATE TABLE, remove corresponding ALTER TABLE statements—schema must be clean.

To restart with fresh DB:
```bash
docker stop backend && docker rm backend && docker volume rm private-tf-runners_backend-data 2>/dev/null; docker-compose up -d backend
```

## Docker Compose

Uses external network `tf-registry-network`. Runner containers use `--network host` with extra_hosts for internal domains.

Run runner manually:
```bash
docker run --rm --network host --add-host=api.runners.lan:127.0.0.1 --env-file .env.runner private-tf-runners-runner
```

## Key Files

| File | Purpose |
|------|---------|
| `cmd/server/main.go` | Backend entry point |
| `cmd/runner/main.go` | Runner entry point |
| `internal/database/database.go` | Schema and migrations |
| `internal/backends/config.go` | Terraform backend schemas |
| `internal/handlers/backend_handler.go` | Backend config validation |
| `internal/models/models.go` | User roles and permissions |
| `frontend/src/lib/permissions.ts` | Frontend permission helpers |
| `frontend/src/hooks/useAuth.tsx` | Auth context with `can()` method |

## Default Credentials

Default admin user seeded on first run:
- Username: `admin`
- Password: `Admin123@Test`