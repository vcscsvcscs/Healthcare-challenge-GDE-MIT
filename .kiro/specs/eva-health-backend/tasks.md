# Implementation Plan: Eva Health Assistant Backend

## Overview

This implementation plan breaks down the Eva Health Assistant backend into discrete, incremental tasks. Each task builds on previous work, with testing integrated throughout. The backend is a Go API that manages conversational health check-ins, stores health data in PostgreSQL, and integrates with Azure services (OpenAI, Speech Service, Blob Storage).

**Infrastructure Prerequisites**: Before starting backend implementation, ensure the following Azure resources are provisioned via Terraform:
- Azure OpenAI service with GPT-4o deployment
- Azure Speech Service (already deployed)
- Azure Blob Storage account with container for health reports and audio recordings
- Use local PostgreSQL for development with docker compose
Fulfilled: 21:37:52.214 STDOUT terraform: Releasing state lock. This may take a few moments...
21:37:52.649 STDOUT terraform: 
21:37:52.649 STDOUT terraform: Apply complete! Resources: 3 added, 1 changed, 0 destroyed.
21:37:52.650 STDOUT terraform: 
21:37:52.650 STDOUT terraform: Outputs:
21:37:52.650 STDOUT terraform: 
21:37:52.651 STDOUT terraform: openai_deployments = {
21:37:52.651 STDOUT terraform:   "gpt-4o" = "/subscriptions/61c53454-ceb0-49ba-bc5a-6178761ee50d/resourceGroups/Solo-1/providers/Microsoft.CognitiveServices/accounts/openai-healthcare-dev/deployments/gpt-4o"
21:37:52.652 STDOUT terraform: }
21:37:52.652 STDOUT terraform: openai_endpoint = "https://openai-healthcare-dev.openai.azure.com/"
21:37:52.652 STDOUT terraform: openai_key = <sensitive>
21:37:52.652 STDOUT terraform: resource_group_name = "Solo-1"
21:37:52.652 STDOUT terraform: speech_endpoint = "https://speech-healthcare-dev.cognitiveservices.azure.com/"
21:37:52.653 STDOUT terraform: speech_key = <sensitive>
21:37:52.653 STDOUT terraform: speech_region = "swedencentral"
21:37:52.653 STDOUT terraform: storage_account_name = "evahealthstoragedev"
21:37:52.653 STDOUT terraform: storage_blob_endpoint = "https://evahealthstoragedev.blob.core.windows.net/"
21:37:52.653 STDOUT terraform: storage_connection_string = <sensitive>
21:37:52.654 STDOUT terraform: storage_containers = [
21:37:52.654 STDOUT terraform:   "audio-recordings",
21:37:52.654 STDOUT terraform:   "health-reports",
21:37:52.654 STDOUT terraform: ]

## Tasks

- [x] 1. Project setup and OpenAPI specification
  - Review existing Go module and update to Go 1.26 if needed
  - Create project structure: internal/, pkg/, integration-tests/ folders in apps/backend/
  - Create /api folder in project root
  - Create OpenAPI 3.0 specification in /api/openapi.json with all endpoints
  - Configure oapi-codegen in pkg/api/ folder
  - Generate initial API types and server interface
  - Set up configuration management with viper
  - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 13.1_

- [ ] 2. Database setup and migrations
  - [ ] 2.1 Update database schema and migrations
    - Review existing migrations in apps/backend/migrations/
    - Create new migrations for all tables (check_in_sessions, conversation_messages, audio_recordings, health_check_ins, medications, medication_logs, menstruation_cycles, blood_pressure_readings, fitness_data, reports)
    - Add indexes for performance optimization
    - Update existing schema if needed
    - _Requirements: 1.1, 1.6, 4.1, 5.1, 6.1_
  
  - [ ]* 2.2 Set up testcontainers for integration tests
    - Configure PostgreSQL testcontainer
    - Create test database setup helper
    - Run migrations in test environment
    - _Requirements: Testing infrastructure_

- [x] 3. Core domain models and repository layer
  - [x] 3.1 Define domain models in pkg/model
    - Create all Go structs (User, Session, Message, HealthCheckIn, Medication, etc.)
    - Add JSON tags for API serialization
    - _Requirements: 1.1, 1.6, 3.2-3.10, 4.1, 5.1, 6.1_
  
  - [x] 3.2 Implement repository layer with pgx
    - Create CheckInRepository with CRUD operations
    - Create MedicationRepository with CRUD operations
    - Create HealthDataRepository for menstruation, blood pressure, fitness data
    - Create DashboardRepository for aggregations
    - Implement proper error handling and context support
    - _Requirements: 1.1, 1.3, 1.6, 4.1-4.6, 5.1-5.5, 6.1-6.5_
  
  - [x] 3.3 Write property tests for repository layer
    - **Property 1: Session Creation Generates Unique IDs**
    - **Property 10: Medication CRUD Preserves ID**
    - **Property 11: Medication Deletion Removes Record**
    - **Property 13: List Sorting Consistency**
    - _Requirements: 1.1, 4.2, 4.3, 4.4, 5.2, 6.5_

- [x] 4. Azure service clients
  - [x] 4.1 Implement Azure OpenAI client
    - Create OpenAIClient wrapper around Azure SDK
    - Implement Complete() method for chat completions
    - Add retry logic with exponential backoff
    - Add logging for token usage and processing time
    - _Requirements: 3.1, 12.3_
  
  - [x] 4.2 Implement Azure Speech Service client
    - Create SpeechServiceClient for speech-to-text
    - Implement StreamAudioToText() for real-time transcription
    - Implement TextToSpeech() for Hungarian voice synthesis
    - Configure Hungarian language (hu-HU) and voice (NoemiNeural)
    - _Requirements: 1.2, 2.2_
  
  - [x] 4.3 Implement Azure Blob Storage client
    - Create BlobStorageClient wrapper
    - Implement UploadPDF() and DownloadPDF()
    - Implement UploadAudio() and DownloadAudio()
    - Add error handling for storage operations
    - _Requirements: 8.3, 8.4_
  
  - [x] 4.4 Write unit tests for Azure clients
    - Mock Azure SDK responses
    - Test error handling and retries
    - Test audio streaming
    - _Requirements: 3.1, 10.6_

- [ ] 5. Checkpoint - Verify infrastructure
  - Ensure all tests pass, ask the user if questions arise.

- [x] 6. Conversation management service
  - [x] 6.1 Implement QuestionFlow
    - Define Hungarian question set (8 questions)
    - Implement GetNextQuestion() and IsComplete()
    - Add question type validation
    - _Requirements: 2.1, 2.2, 2.4_
  
  - [x] 6.2 Implement DataExtractor
    - Create AI prompt template for data extraction
    - Implement Extract() method using Azure OpenAI
    - Parse JSON response into ExtractedData struct
    - Handle extraction failures gracefully
    - _Requirements: 3.1-3.12_
  
  - [x] 6.3 Implement CheckInService
    - Implement StartSession() with audio generation
    - Implement StreamAudioToSpeech() for real-time transcription
    - Implement ProcessResponse() with conversation state management
    - Implement GetQuestionAudio() with caching
    - Implement CompleteSession() with data extraction
    - Implement GetSessionStatus()
    - Add session timeout logic (30 minutes)
    - _Requirements: 1.1-1.7, 2.1-2.6, 3.1-3.12_
  
  - [x] 6.4 Write property tests for conversation flow
    - **Property 2: Session Creation Returns First Question**
    - **Property 3: Response Storage and Progression**
    - **Property 4: Session Completion After All Questions**
    - **Property 5: Data Extraction Triggers on Completion**
    - **Property 6: Session Timeout After Inactivity**
    - **Property 7: Conversation Time Limit**
    - **Property 8: Data Extraction Output Structure**
    - **Property 9: AI Failure Fallback**
    - _Requirements: 1.1-1.7, 2.1-2.6, 3.1-3.12_

- [x] 7. Health data management services
  - [x] 7.1 Implement MedicationService
    - Implement AddMedication(), ListMedications(), UpdateMedication(), DeleteMedication()
    - Add medication adherence logging
    - Handle inactive medications (past end date)
    - _Requirements: 4.1-4.6_
  
  - [x] 7.2 Implement HealthDataService
    - Implement LogMenstruation() and GetMenstruationHistory()
    - Implement LogBloodPressure() and GetBloodPressureHistory()
    - Implement SyncFitnessData() with deduplication
    - Implement GetFitnessHistory()
    - Add input validation for all health metrics
    - _Requirements: 5.1-5.5, 6.1-6.5_
  
  - [x] 7.3 Write property tests for health data services
    - **Property 12: Inactive Medication Retention**
    - **Property 14: Input Validation Rejects Invalid Ranges**
    - **Property 15: Enum Validation**
    - _Requirements: 4.5, 5.3, 6.2-6.4_

- [x] 8. Dashboard and reporting services
  - [x] 8.1 Implement DashboardService
    - Implement GetSummary() with time range filtering
    - Implement GetTrends() with aggregations (average pain, mood distribution, energy levels)
    - Implement time-series data grouping by date
    - Handle empty datasets gracefully
    - _Requirements: 7.1-7.5_
  
  - [x] 8.2 Implement PDF generation
    - Create PDFGenerator using go-pdf/fpdf
    - Implement Generate() with all report sections
    - Format report professionally for medical use
    - Include charts/graphs for trends
    - _Requirements: 8.1, 8.2, 8.5_
  
  - [x] 8.3 Implement ReportService
    - Implement GenerateReport() asynchronously
    - Implement GetReport() for PDF download
    - Store reports in Azure Blob Storage
    - Create report records in database
    - _Requirements: 8.1-8.6_
  
  - [x] 8.4 Write property tests for dashboard and reports
    - **Property 16: Dashboard Time Range Filtering**
    - **Property 17: Dashboard Aggregation Accuracy**
    - **Property 18: Time Series Data Grouping**
    - **Property 19: Report Content Completeness**
    - **Property 20: Report Storage and Retrieval Round Trip**
    - _Requirements: 7.2-7.4, 8.1-8.4_

- [ ] 9. Checkpoint - Verify business logic
  - Ensure all tests pass, ask the user if questions arise.

- [x] 10. API handlers implementation
  - [x] 10.1 Implement CheckInHandler
    - Implement PostApiV1CheckinStart()
    - Implement PostApiV1CheckinAudioStream() with WebSocket support
    - Implement PostApiV1CheckinRespond()
    - Implement GetApiV1CheckinStatusSessionId()
    - Implement GetApiV1CheckinQuestionAudioSessionIdQuestionId()
    - Implement PostApiV1CheckinComplete()
    - Add request validation and error handling
    - _Requirements: 1.1-1.7, 2.1-2.6, 11.1-11.6_
  
  - [x] 10.2 Implement MedicationHandler
    - Implement POST, GET, PUT, DELETE endpoints
    - Add request validation
    - _Requirements: 4.1-4.6, 11.1-11.6_
  
  - [x] 10.3 Implement HealthHandler
    - Implement menstruation endpoints
    - Implement blood pressure endpoints
    - Implement fitness sync endpoint
    - Add input validation for all health metrics
    - _Requirements: 5.1-5.5, 6.1-6.5, 11.1-11.6_
  
  - [x] 10.4 Implement DashboardHandler
    - Implement GetSummary() and GetTrends()
    - Add time range parameter handling
    - _Requirements: 7.1-7.5, 11.1-11.6_
  
  - [x] 10.5 Implement ReportHandler
    - Implement GenerateReport() and GetReport()
    - Handle async report generation
    - _Requirements: 8.1-8.6, 11.1-11.6_
  
  - [x] 10.6 Write property tests for error handling
    - **Property 24: Error Response Structure**
    - **Property 25: Request Validation Completeness**
    - _Requirements: 11.1-11.6_

- [x] 11. Logging, monitoring, and security
  - [x] 11.1 Implement structured logging with zap
    - Set up zap logger with JSON output
    - Add request logging middleware
    - Add error logging with stack traces
    - Add AI operation logging (processing time, token usage)
    - Add session completion logging
    - _Requirements: 12.1-12.5_
  
  - [x] 11.2 Implement audit logging
    - Create audit log entries for all data modifications
    - Log user ID, operation type, timestamp, affected resource
    - _Requirements: 10.5_
  
  - [x] 11.3 Implement data encryption and security
    - Add AES-256 encryption for sensitive health data fields
    - Implement data deletion (GDPR compliance)
    - Implement data export to JSON
    - _Requirements: 10.1-10.5_
  
  - [x] 11.4 Write property tests for logging and security
    - **Property 21: Data Deletion Completeness**
    - **Property 22: Data Export Completeness**
    - **Property 23: Audit Log Creation**
    - **Property 26: Request Logging**
    - **Property 27: Error Logging Detail**
    - **Property 28: AI Operation Logging**
    - **Property 29: Session Completion Logging**
    - _Requirements: 10.3-10.5, 12.1-12.4_

- [x] 12. Server setup and middleware
  - [x] 12.1 Update main server application in apps/backend/main.go
    - Initialize Gin router
    - Register generated API handlers
    - Add recovery middleware
    - Add CORS middleware
    - Add request logging middleware
    - Add tracing middleware
    - _Requirements: 9.1, 12.1_
  
  - [x] 12.2 Implement health check endpoint
    - Add /health endpoint
    - Check database connectivity
    - Return service status
    - _Requirements: Deployment_
  
  - [x] 12.3 Implement graceful shutdown
    - Handle SIGINT and SIGTERM signals
    - Close database connections
    - Drain in-flight requests
    - _Requirements: Deployment_

- [x] 13. Integration tests
  - [x] 13.1 Write check-in flow integration test
    - Test complete check-in flow from start to completion
    - Verify audio streaming and transcription
    - Verify data extraction and storage
    - _Requirements: 1.1-1.7, 2.1-2.6, 3.1-3.12_
  
  - [x] 13.2 Write medication management integration test
    - Test CRUD operations
    - Test medication adherence logging
    - _Requirements: 4.1-4.6_
  
  - [x] 13.3 Write health data tracking integration test
    - Test menstruation, blood pressure, fitness data
    - Test data retrieval and filtering
    - _Requirements: 5.1-5.5, 6.1-6.5_
  
  - [x] 13.4 Write dashboard and reporting integration test
    - Test dashboard aggregations
    - Test report generation and download
    - _Requirements: 7.1-7.5, 8.1-8.6_

- [ ] 14. Final checkpoint and documentation
  - Ensure all tests pass (unit, property, integration)
  - Verify OpenAPI documentation is complete and accurate
  - Test all endpoints manually or with Postman
  - Create .env.example file with all configuration variables
  - Document deployment steps
  - Ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional test tasks that can be skipped for faster MVP
- Each task references specific requirements for traceability
- Property tests validate universal correctness properties using gopter (minimum 100 iterations)
- Unit tests validate specific examples and edge cases
- Integration tests use testcontainers for isolated PostgreSQL instances
- The OpenAPI specification drives code generation, ensuring type safety
- All Azure service calls include retry logic and proper error handling
- Logging uses zap for structured JSON output
- The project follows Go best practices with clear separation of concerns
