# Healthcare Backend

Go backend service for the Healthcare application.

## Tech Stack

- **Go 1.25+** - Programming language
- **Gin** - HTTP web framework
- **Zap** - Structured logging
- **Docker** - Containerization
- **Azure Container Registry** - Container registry

## Development

### Prerequisites
- Go 1.25+
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

## API Endpoints

- `GET /` - Root endpoint (returns service info)
- `GET /health` - Health check endpoint
- `GET /api/v1/status` - Status endpoint

## Environment Variables

- `PORT` - Server port (default: 8080)
- `ENV` - Environment mode (`production` or `development`, affects logging)
