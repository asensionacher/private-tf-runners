# Private TF Runners

Self-hosted Terraform/OpenTofu deployment platform with a web UI for managing stacks, runs, and runners.

## Architecture

- **Backend**: Go + Gin framework (port 8080)
- **Frontend**: React + TypeScript + TailwindCSS
- **Database**: SQLite (embedded)
- **Reverse Proxy**: Nginx (SSL termination)
- **Security**: JWT authentication, CSRF protection, role-based access control

## Quick Start

```bash
# Create external network
docker network create tf-registry-network

# Start all services
docker-compose up -d

# Access the UI at https://runners.lan
```

## Default Credentials

On first run, a default admin user is created:
- **Username**: `admin`
- **Password**: `Admin123@Test`

## User Roles

| Role | Permissions |
|------|-------------|
| **Viewer** | Read stacks, runs, runners |
| **Operator** | Create/delete stacks and runs, manage runners, reset runner tokens |
| **Admin** | All operator permissions + user management |

## Services

| Service | Port | Description |
|---------|------|-------------|
| nginx | 443 | Reverse proxy with SSL |
| backend | 8080 | REST API server |
| frontend | 80 | Web UI |

## Configuration

### Environment Variables (.env)

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_HOST` | Backend bind address | `0.0.0.0` |
| `SERVER_PORT` | Backend port | `8080` |
| `DATABASE_PATH` | SQLite database path | `./data/runners.db` |
| `JWT_SECRET` | JWT signing secret | (required) |
| `ENCRYPTION_KEY` | Data encryption key | (required) |

### Terraform Backend Support

Supported Terraform state backends for runs:
- `azurerm` - Azure Blob Storage
- `s3` - Amazon S3 (with assume_role support)
- `http` - Generic HTTP REST endpoint
- `pg` - PostgreSQL
- `kubernetes` - Kubernetes secrets
- `gcs` - Google Cloud Storage

## API Endpoints

### Authentication
- `POST /api/auth/login` - Login
- `POST /api/auth/logout` - Logout
- `GET /api/auth/me` - Current user info

### Stacks
- `GET /api/stacks` - List stacks
- `POST /api/stacks` - Create stack
- `GET /api/stacks/:id` - Get stack
- `PUT /api/stacks/:id` - Update stack
- `DELETE /api/stacks/:id` - Delete stack

### Runs
- `GET /api/runs` - List runs
- `POST /api/runs` - Create run
- `GET /api/runs/:id` - Get run details

### Runners
- `GET /api/runners` - List runners
- `POST /api/runners` - Create runner
- `POST /api/runners/:id/reset-token` - Reset runner token

### Users (Admin only)
- `GET /api/users` - List users
- `POST /api/users` - Create user
- `POST /api/users/:id/reset-password` - Reset user password
- `DELETE /api/users/:id` - Delete user

## Docker Commands

```bash
# Build images
docker build -t private-tf-runners-backend -f ./Dockerfile .
docker build -t private-tf-runners-runner --no-cache -f ./cmd/runner/Dockerfile .

# Restart with fresh database
docker stop backend && docker rm backend && docker volume rm private-tf-runners_backend-data
docker-compose up -d backend

# Run runner manually
docker run --rm --network host --add-host=api.runners.lan:127.0.0.1 --env-file .env.runner private-tf-runners-runner
```

## Project Structure

```
├── cmd/
│   ├── server/main.go      # Backend entry point
│   └── runner/main.go       # Runner entry point
├── internal/
│   ├── database/            # SQLite database layer
│   ├── handlers/           # HTTP request handlers
│   ├── middleware/         # Auth, rate limiting, security
│   ├── models/             # Data models and permissions
│   └── backends/           # Terraform backend schemas
├── frontend/               # React application
│   └── src/
│       ├── pages/          # Page components
│       ├── components/     # Reusable components
│       ├── hooks/          # React hooks (useAuth)
│       └── lib/             # API client, permissions
├── nginx/                  # Nginx configuration
├── docker-compose.yml      # Service orchestration
└── Dockerfile              # Backend container
```

## Security Features

- JWT authentication with short expiration
- CSRF protection on state-changing operations
- Security headers (CSP, X-Frame-Options, etc.)
- SQL injection prevention (parameterized queries)
- Rate limiting on authentication endpoints
- bcrypt password hashing
- Audit logging for sensitive operations