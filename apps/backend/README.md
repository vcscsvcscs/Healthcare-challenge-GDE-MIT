# Eva Health Assistant Backend

Voice-first health data aggregation platform backend built in Go. Manages conversational health check-ins, stores structured health data in PostgreSQL, and generates professional medical reports.

## Architecture

- **Language**: Go 1.26+
- **Web Framework**: Gin (via oapi-codegen generated handlers)
- **Database**: PostgreSQL 15+ with pgx driver
- **API Specification**: OpenAPI 3.0 with oapi-codegen code generation
- **Azure Services**: Azure OpenAI, Azure Speech Service, Azure Blob Storage
- **Configuration**: Environment variables with viper

## Project Structure

```
apps/backend/
├── internal/           # Private application code
│   └── config/        # Configuration management
├── pkg/               # Public packages
│   └── api/          # Generated API types and handlers
├── integration-tests/ # Integration tests
├── migrations/        # Database migrations
└── main.go           # Application entry point
```

## Setup

### Prerequisites

- Go 1.26+
- PostgreSQL 15+
- Azure OpenAI service
- Azure Speech Service
- Azure Blob Storage

### Configuration

Copy `.env.example` to `.env` and fill in your Azure credentials:

```bash
cp .env.example .env
```

Required environment variables:
- `DATABASE_URL`: PostgreSQL connection string
- `AZURE_OPENAI_ENDPOINT`: Azure OpenAI endpoint
- `AZURE_OPENAI_API_KEY`: Azure OpenAI API key
- `AZURE_OPENAI_DEPLOYMENT`: Azure OpenAI deployment name (e.g., gpt-4o)
- `AZURE_SPEECH_KEY`: Azure Speech Service subscription key
- `AZURE_SPEECH_REGION`: Azure Speech Service region
- `AZURE_STORAGE_CONNECTION_STRING`: Azure Blob Storage connection string

### Install Dependencies

```bash
go mod download
```

### Generate API Code

The API code is generated from the OpenAPI specification using oapi-codegen:

```bash
go generate ./pkg/api
```

### Run Database Migrations

```bash
# Install golang-migrate if not already installed
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path migrations -database "${DATABASE_URL}" up
```

### Run the Server

```bash
go run main.go
```

The server will start on port 8080 (or the port specified in the `PORT` environment variable).

## API Documentation

The API is documented using OpenAPI 3.0 specification in `/api/openapi.json`.

Key endpoints:
- `POST /api/v1/checkin/start` - Start new check-in session
- `POST /api/v1/checkin/audio-stream` - Stream audio for transcription
- `POST /api/v1/checkin/respond` - Submit user response
- `POST /api/v1/checkin/complete` - Complete check-in session
- `POST /api/v1/health/medications` - Add medication
- `GET /api/v1/health/medications` - List medications
- `POST /api/v1/health/menstruation` - Log menstruation data
- `POST /api/v1/health/blood-pressure` - Log blood pressure
- `GET /api/v1/dashboard/summary` - Get dashboard summary
- `POST /api/v1/reports/generate` - Generate health report

## Development

### Code Generation

When the OpenAPI specification is updated, regenerate the API code:

```bash
go generate ./pkg/api
```

### Testing

Run tests:

```bash
go test ./...
```

Run integration tests:

```bash
go test ./integration-tests/...
```

## Tech Stack

- **Go 1.26+** - Programming language
- **Gin** - HTTP web framework
- **Zap** - Structured logging
- **Viper** - Configuration management
- **oapi-codegen** - OpenAPI code generation
- **Docker** - Containerization
- **Azure Container Registry** - Container registry

## Development

### Prerequisites
- Go 1.26+
- Docker
- Azure CLI
- mise (for task automation)

### Local Development

```bash
# Run locally
mise run dev:backend

# Or directly with Go
go run main.go
```

## Docker

### Build Image

```bash
# Using mise
mise run docker:build:backend

# Or directly
docker build -t backend:latest .
```

### Run Container

```bash
docker run -p 8080:8080 backend:latest
```

## Azure Container Registry

### Build and Push to ACR

```bash
# Complete workflow (build, tag, push)
mise run docker:build-push:backend

# Or step by step
mise run docker:build:backend
mise run docker:tag:backend
mise run docker:push:backend
```

### ACR Management

```bash
# Login to ACR
mise run acr:login

# List all repositories
mise run acr:list

# Show backend image tags
mise run acr:show:backend
```

## Terraform Automation

The infrastructure setup includes automatic Docker build and push:

```bash
cd infra

# Initialize Terraform
terraform init

# Plan (will show Docker build in plan)
terraform plan

# Apply (will build and push Docker image)
terraform apply
```

The Docker image is automatically built and pushed when:
- Dockerfile changes
- go.mod changes
- ACR configuration changes

## Deployment

See deployment documentation for production setup instructions.
