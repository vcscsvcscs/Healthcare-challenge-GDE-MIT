# Integration Tests

This directory contains integration tests for the Eva Health Assistant backend. These tests verify the complete functionality of the system including database operations, Azure service integrations, and end-to-end API flows.

## Prerequisites

### Database Setup

The integration tests require a PostgreSQL database. You can use Docker to run a test database:

```bash
docker run -d \
  --name eva-health-test-db \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=eva_health_test \
  -p 5432:5432 \
  postgres:15
```

Run migrations on the test database:

```bash
cd apps/backend
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/eva_health_test?sslmode=disable"
migrate -path migrations -database $DATABASE_URL up
```

### Azure Services (Optional)

By default, the tests use mock Azure clients. To test with real Azure services, set the following environment variables:

```bash
export USE_REAL_AZURE=true
export AZURE_OPENAI_ENDPOINT="https://your-openai.openai.azure.com/"
export AZURE_OPENAI_KEY="your-openai-key"
export AZURE_OPENAI_DEPLOYMENT="gpt-4o"
export AZURE_SPEECH_KEY="your-speech-key"
export AZURE_SPEECH_REGION="swedencentral"
export AZURE_STORAGE_ACCOUNT_NAME="your-storage-account"
export AZURE_STORAGE_ACCOUNT_KEY="your-storage-key"
export AZURE_STORAGE_CONTAINER="health-reports"
```

## Running the Tests

### Using mise (Recommended)

```bash
cd apps/backend

# Install dependencies and tools
mise run install

# Run integration tests with mock Azure services
mise run test-integration

# Run integration tests with real Azure services (requires env vars)
mise run test-integration-real

# Run unit tests only
mise run test

# Run unit tests with coverage
mise run test-coverage

# Reset test database
mise run db-reset

# See all available tasks
mise tasks
```

### Using the shell script

```bash
cd apps/backend/integration-tests

# Run with mock Azure services
./run-tests.sh

# Run with real Azure services
./run-tests.sh --real-azure

# Stop database
./run-tests.sh --stop

# Clean up everything
./run-tests.sh --cleanup
```

### Manual execution

```bash
cd apps/backend

# Run all integration tests
go test -v ./integration-tests/...

# Run with mock Azure services (default)
export TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/eva_health_test?sslmode=disable"
go test -v ./integration-tests/...

# Run with real Azure services
export TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/eva_health_test?sslmode=disable"
export USE_REAL_AZURE=true
export AZURE_OPENAI_ENDPOINT="..."
export AZURE_OPENAI_KEY="..."
export AZURE_OPENAI_DEPLOYMENT="gpt-4o"
export AZURE_SPEECH_KEY="..."
export AZURE_SPEECH_REGION="swedencentral"
export AZURE_STORAGE_ACCOUNT_NAME="..."
export AZURE_STORAGE_ACCOUNT_KEY="..."
export AZURE_STORAGE_CONTAINER="health-reports"
go test -v ./integration-tests/...

# Skip integration tests (for unit test runs)
go test -short ./...
```

## Test Coverage

### Check-in Flow Integration Test

**File**: `checkin_flow_test.go`

**Requirements Covered**: 1.1-1.7, 2.1-2.6, 3.1-3.12

**Test Scenarios**:

1. **Complete check-in flow**
   - Start a new check-in session
   - Verify first question is returned in Hungarian
   - Answer all 8 questions in the conversation flow
   - Complete the session and trigger data extraction
   - Verify extracted data structure and validity
   - Verify data persistence in database

2. **Audio streaming and transcription**
   - Generate test audio using Text-to-Speech
   - Stream audio to the API
   - Verify transcription is returned
   - Verify audio is processed correctly

3. **Session timeout handling**
   - Covered in unit tests (requires time manipulation)

## Test Data

The integration tests use realistic Hungarian responses to simulate actual user interactions:

- General feeling: "Jól érzem magam ma, kicsit fáradt vagyok."
- Physical activity: "Igen, reggel futottam 5 kilométert."
- Meals: "Reggelire zabkását ettem, ebédre csirkét rizzsel, vacsorára salátát."
- Pain: "Igen, kicsit fáj a fejem."
- Sleep: "Jól aludtam, 8 órát."
- Energy: "Közepes az energiaszintem."
- Medication: "Igen, beszedtem minden gyógyszeremet."
- Additional notes: "Semmi különös, minden rendben."

## Troubleshooting

### Database connection errors

Ensure PostgreSQL is running and the database exists:

```bash
docker ps | grep eva-health-test-db
psql -h localhost -U postgres -d eva_health_test -c "SELECT 1;"
```

### Azure service errors

When using real Azure services, verify your credentials:

```bash
# Test OpenAI connection
curl -X POST "$AZURE_OPENAI_ENDPOINT/openai/deployments/$AZURE_OPENAI_DEPLOYMENT/chat/completions?api-version=2024-08-01-preview" \
  -H "api-key: $AZURE_OPENAI_KEY" \
  -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"test"}]}'

# Test Speech Service connection
curl -X POST "https://$AZURE_SPEECH_REGION.stt.speech.microsoft.com/speech/recognition/conversation/cognitiveservices/v1?language=hu-HU" \
  -H "Ocp-Apim-Subscription-Key: $AZURE_SPEECH_KEY" \
  -H "Content-Type: audio/wav" \
  --data-binary @test.wav
```

### Migration errors

If migrations fail, reset the test database:

```bash
docker stop eva-health-test-db
docker rm eva-health-test-db
# Then recreate and run migrations again
```

## CI/CD Integration

For CI/CD pipelines, use mock Azure services and a containerized database:

```yaml
# Example GitHub Actions workflow
- name: Start PostgreSQL
  run: |
    docker run -d \
      --name postgres \
      -e POSTGRES_PASSWORD=postgres \
      -e POSTGRES_DB=eva_health_test \
      -p 5432:5432 \
      postgres:15

- name: Run migrations
  run: |
    cd apps/backend
    migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/eva_health_test?sslmode=disable" up

- name: Run integration tests
  run: |
    cd apps/backend
    export TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/eva_health_test?sslmode=disable"
    go test -v ./integration-tests/...
```

## Future Tests

Additional integration tests to be added:

- Medication management flow (Task 13.2)
- Health data tracking flow (Task 13.3)
- Dashboard and reporting flow (Task 13.4)
