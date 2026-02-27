# Healthcare Challenge - GDE MIT

A healthcare application with speech-to-text capabilities using Azure Cognitive Services, built with Go backend and SvelteKit frontend.

## Architecture

- **Backend**: Go with Gin framework, PostgreSQL database
- **Frontend**: SvelteKit with TypeScript
- **Infrastructure**: Azure Speech Service (public access)
- **Database**: PostgreSQL with golang-migrate
- **Containerization**: Docker Compose for local development

## Prerequisites

- Go 1.25+
- Node.js (latest)
- Docker & Docker Compose
- Azure CLI
- Terraform/Terragrunt
- mise (task runner)

## Quick Start

### 1. Clone and Setup

```bash
# Install dependencies
mise install

# Setup projects
mise run setup
```

### 2. Configure Azure Speech Service

```bash
# Login to Azure
az login

# Deploy infrastructure (creates Speech service)
cd infra
terragrunt plan
terragrunt apply

# Get Speech service credentials
terragrunt output speech_key
terragrunt output speech_region
```

### 3. Configure Environment

```bash
# Copy example env file
cp .env.example .env

# Edit .env and add your Azure Speech credentials
# AZURE_SPEECH_KEY=your_key_here
# AZURE_SPEECH_REGION=swedencentral
```

### 4. Start Services

```bash
# Start all services (PostgreSQL, migrations, backend)
mise run compose:up

# Or start individually
mise run db:up          # Start PostgreSQL
mise run db:migrate     # Run migrations
mise run dev:backend    # Start backend
```

### 5. Verify Setup

```bash
# Check health
curl http://localhost:8080/health

# View logs
mise run compose:logs
```

## Development

### Database Management

```bash
# Start database
mise run db:up

# Run migrations
mise run db:migrate

# Create new migration
mise run db:migrate:create add_new_table

# Rollback migration
mise run db:migrate:down

# Connect to database
mise run db:psql
```

### Backend Development

```bash
# Run locally (requires DATABASE_URL)
mise run dev:backend

# Or with Docker
docker compose up backend

# Build Docker image
mise run docker:build:backend
```

### Frontend Development

```bash
# Install dependencies
cd apps/frontend
pnpm install

# Run dev server
pnpm run dev

# Build for production
pnpm run build
```

## Docker Compose Services

- **postgres**: PostgreSQL 16 database
- **migrate**: Database migration runner (golang-migrate)
- **backend**: Go API server

## Database Schema

### Tables

- **users**: User accounts
- **sessions**: User sessions
- **transcriptions**: Speech-to-text transcriptions

See `apps/backend/migrations/` for full schema.

## API Endpoints

- `GET /` - Service info
- `GET /health` - Health check (includes DB status)
- `GET /api/v1/status` - Service status
- `GET /api/v1/users` - List users

## Infrastructure

The infrastructure is managed with Terraform/Terragrunt and includes:

- Azure Speech Service (public access, F0/S0 tier)
- Resource Group: Solo-1
- Region: Sweden Central

### Deploy Infrastructure

```bash
cd infra
terragrunt plan
terragrunt apply
```

### Get Credentials

```bash
# Speech service key
terragrunt output -raw speech_key

# Speech service endpoint
terragrunt output speech_endpoint
```

## Mise Tasks Reference

### Setup & Development
- `mise run setup` - Initial project setup
- `mise run dev` - Run frontend and backend
- `mise run dev:backend` - Run backend only
- `mise run dev:frontend` - Run frontend only

### Database
- `mise run db:up` - Start PostgreSQL
- `mise run db:migrate` - Run migrations
- `mise run db:migrate:create <name>` - Create migration
- `mise run db:migrate:down` - Rollback migration
- `mise run db:psql` - Connect to database
- `mise run db:down` - Stop database

### Docker Compose
- `mise run compose:up` - Start all services
- `mise run compose:down` - Stop all services
- `mise run compose:logs` - View logs
- `mise run compose:rebuild` - Rebuild and restart

### Infrastructure
- `mise run tf:plan` - Terraform plan
- `mise run tf:apply` - Terraform apply
- `mise run azure:login` - Login to Azure
- `mise run azure:whoami` - Show Azure account

## Project Structure

```
.
├── apps/
│   ├── backend/          # Go backend
│   │   ├── migrations/   # Database migrations
│   │   ├── main.go       # Application entry point
│   │   └── Dockerfile    # Backend container
│   ├── frontend/         # SvelteKit frontend
│   └── mobile/           # Mobile app (future)
├── infra/                # Terraform infrastructure
│   ├── modules/          # Terraform modules
│   └── main.tf           # Main infrastructure config
├── docker-compose.yaml   # Local development services
├── mise.toml             # Task automation
└── README.md             # This file
```

## Environment Variables

### Backend
- `PORT` - Server port (default: 8080)
- `ENV` - Environment (development/production)
- `DATABASE_URL` - PostgreSQL connection string
- `AZURE_SPEECH_KEY` - Azure Speech service key
- `AZURE_SPEECH_REGION` - Azure region

## Troubleshooting

### Database Connection Issues

```bash
# Check if PostgreSQL is running
docker compose ps postgres

# View database logs
docker compose logs postgres

# Restart database
mise run db:down
mise run db:up
```

### Migration Issues

```bash
# Check migration status
docker compose logs migrate

# Force migration version
docker compose run --rm migrate \
  -path /migrations \
  -database "postgres://healthcare_user:healthcare_pass@postgres:5432/healthcare?sslmode=disable" \
  force <version>
```

### Backend Issues

```bash
# Check backend logs
docker compose logs backend

# Rebuild backend
docker compose up -d --build backend
```

## License

MIT
