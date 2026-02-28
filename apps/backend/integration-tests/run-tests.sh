#!/bin/bash

# Eva Health Backend Integration Test Runner
# This script sets up the test environment and runs integration tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
DB_CONTAINER_NAME="eva-health-test-db"
DB_NAME="eva_health_test"
DB_USER="postgres"
DB_PASSWORD="postgres"
DB_PORT="5432"
TEST_DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@localhost:${DB_PORT}/${DB_NAME}?sslmode=disable"

# Function to print colored messages
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
    print_info "Docker is running"
}

# Function to start test database
start_database() {
    print_info "Starting test database..."
    
    # Check if container already exists
    if docker ps -a --format '{{.Names}}' | grep -q "^${DB_CONTAINER_NAME}$"; then
        print_warn "Database container already exists"
        
        # Check if it's running
        if docker ps --format '{{.Names}}' | grep -q "^${DB_CONTAINER_NAME}$"; then
            print_info "Database container is already running"
        else
            print_info "Starting existing database container..."
            docker start ${DB_CONTAINER_NAME}
        fi
    else
        print_info "Creating new database container..."
        docker run -d \
            --name ${DB_CONTAINER_NAME} \
            -e POSTGRES_PASSWORD=${DB_PASSWORD} \
            -e POSTGRES_DB=${DB_NAME} \
            -p ${DB_PORT}:5432 \
            postgres:15
    fi
    
    # Wait for database to be ready
    print_info "Waiting for database to be ready..."
    sleep 3
    
    max_attempts=10
    attempt=0
    while [ $attempt -lt $max_attempts ]; do
        if docker exec ${DB_CONTAINER_NAME} pg_isready -U ${DB_USER} > /dev/null 2>&1; then
            print_info "Database is ready"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 1
    done
    
    print_error "Database failed to start"
    exit 1
}

# Function to run migrations
run_migrations() {
    print_info "Running database migrations..."
    
    cd ..
    
    # Check if migrate tool is installed
    if ! command -v migrate &> /dev/null; then
        print_error "migrate tool is not installed"
        print_info "Install it with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
        exit 1
    fi
    
    migrate -path migrations -database "${TEST_DATABASE_URL}" up
    
    print_info "Migrations completed"
    
    cd integration-tests
}

# Function to run tests
run_tests() {
    local use_real_azure=$1
    
    cd ..
    
    if [ "$use_real_azure" = "true" ]; then
        print_info "Running integration tests with REAL Azure services..."
        
        # Check required environment variables
        if [ -z "$AZURE_OPENAI_ENDPOINT" ]; then
            print_error "AZURE_OPENAI_ENDPOINT is not set"
            exit 1
        fi
        if [ -z "$AZURE_OPENAI_KEY" ]; then
            print_error "AZURE_OPENAI_KEY is not set"
            exit 1
        fi
        if [ -z "$AZURE_OPENAI_DEPLOYMENT" ]; then
            print_error "AZURE_OPENAI_DEPLOYMENT is not set"
            exit 1
        fi
        if [ -z "$AZURE_SPEECH_KEY" ]; then
            print_error "AZURE_SPEECH_KEY is not set"
            exit 1
        fi
        if [ -z "$AZURE_SPEECH_REGION" ]; then
            print_error "AZURE_SPEECH_REGION is not set"
            exit 1
        fi
        
        TEST_DATABASE_URL="${TEST_DATABASE_URL}" \
        USE_REAL_AZURE=true \
        go test -v ./integration-tests/...
    else
        print_info "Running integration tests with MOCK Azure services..."
        
        TEST_DATABASE_URL="${TEST_DATABASE_URL}" \
        go test -v ./integration-tests/...
    fi
    
    cd integration-tests
}

# Function to stop database
stop_database() {
    print_info "Stopping test database..."
    docker stop ${DB_CONTAINER_NAME} > /dev/null 2>&1 || true
    print_info "Database stopped"
}

# Function to clean up
cleanup() {
    print_info "Cleaning up..."
    docker stop ${DB_CONTAINER_NAME} > /dev/null 2>&1 || true
    docker rm ${DB_CONTAINER_NAME} > /dev/null 2>&1 || true
    print_info "Cleanup complete"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --real-azure    Run tests with real Azure services (requires env vars)"
    echo "  --cleanup       Stop and remove test database"
    echo "  --stop          Stop test database (keep container)"
    echo "  --help          Show this help message"
    echo ""
    echo "Environment variables for --real-azure:"
    echo "  AZURE_OPENAI_ENDPOINT"
    echo "  AZURE_OPENAI_KEY"
    echo "  AZURE_OPENAI_DEPLOYMENT"
    echo "  AZURE_SPEECH_KEY"
    echo "  AZURE_SPEECH_REGION"
    echo "  AZURE_STORAGE_ACCOUNT_NAME"
    echo "  AZURE_STORAGE_ACCOUNT_KEY"
    echo "  AZURE_STORAGE_CONTAINER"
}

# Main script
main() {
    local use_real_azure="false"
    local action="test"
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --real-azure)
                use_real_azure="true"
                shift
                ;;
            --cleanup)
                action="cleanup"
                shift
                ;;
            --stop)
                action="stop"
                shift
                ;;
            --help)
                show_usage
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # Execute action
    case $action in
        test)
            check_docker
            start_database
            run_migrations
            run_tests "$use_real_azure"
            print_info "Tests completed successfully!"
            ;;
        stop)
            stop_database
            ;;
        cleanup)
            cleanup
            ;;
    esac
}

# Run main function
main "$@"
