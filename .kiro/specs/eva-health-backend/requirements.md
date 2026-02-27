# Requirements Document: Eva Health Assistant Backend

## Introduction

Eva Health Assistant is a voice-first health data aggregation platform backend built in Go. The system enables Hungarian-speaking users to conduct conversational health check-ins, stores structured health data in PostgreSQL, and generates professional medical reports for doctor consultations. The backend integrates with Azure Speech Service for speech-to-text conversion and Azure OpenAI for conversation management and data extraction.

## Glossary

- **Eva_System**: The complete backend API system that manages health check-ins, data storage, and report generation
- **Check_In_Session**: A conversational interaction between the user and Eva where health questions are asked and answered
- **Health_Data_Store**: PostgreSQL database containing all user health records
- **Conversation_Manager**: Component that manages the flow of questions and responses during check-ins
- **Data_Extractor**: AI-powered component that converts conversational responses into structured health data
- **Report_Generator**: Component that creates PDF medical reports from stored health data
- **Mobile_Client**: Flutter Android application that users interact with
- **Azure_Speech_Service**: Microsoft Azure service for speech-to-text conversion (already deployed)
- **Azure_OpenAI**: Microsoft Azure service for AI-powered conversation and data extraction
- **User**: Hungarian-speaking individual using the Eva Health Assistant mobile app
- **Doctor_Report**: PDF document containing comprehensive health data for medical consultation
- **Medication_Adherence**: Boolean or partial indicator of whether user took prescribed medications
- **Health_Metric**: Quantifiable health measurement (pain level, mood, energy level, blood pressure)
- **Menstruation_Cycle**: Time period tracking menstrual cycle with associated symptoms and flow intensity
- **Session_State**: Current status and conversation history of an active check-in session

## Requirements

### Requirement 1: Conversational Health Check-In Management

**User Story:** As a user, I want to complete daily health check-ins through a conversational interface, so that I can easily track my health without manual data entry.

#### Acceptance Criteria

1. WHEN a user initiates a check-in, THE Eva_System SHALL create a new Check_In_Session with unique session ID and timestamp
2. WHEN a Check_In_Session is created, THE Conversation_Manager SHALL present the first health question in Hungarian
3. WHEN a user submits a response, THE Eva_System SHALL store the response in the conversation history and advance to the next question
4. WHEN all questions are answered, THE Eva_System SHALL mark the Check_In_Session as complete
5. WHEN a Check_In_Session is marked complete, THE Data_Extractor SHALL process the conversation history and extract structured health data
6. WHEN structured data is extracted, THE Eva_System SHALL store the data in the Health_Data_Store with the current date and user ID
7. IF a Check_In_Session is inactive for more than 30 minutes, THEN THE Eva_System SHALL mark the session as expired

### Requirement 2: Health Question Flow

**User Story:** As a user, I want Eva to ask me relevant health questions in Hungarian in a natural conversational manner, so that I can provide comprehensive health information without feeling like I'm filling out a form.

#### Acceptance Criteria

1. THE Conversation_Manager SHALL ask questions in a natural conversational flow covering: general feeling about the day, physical activity (sports, walks), meals (breakfast, lunch, dinner), pain or discomfort, sleep quality, energy level, medication adherence, and additional notes
2. WHEN presenting a question, THE Eva_System SHALL return the question text in Hungarian
3. WHEN a user provides an answer, THE Eva_System SHALL validate that the response is not empty before proceeding
4. WHEN asking about pain, THE Conversation_Manager SHALL use conversational phrasing such as "Fáj valami ma?" (Does anything hurt today?) and follow up based on the response
5. WHEN all mandatory questions are answered, THE Conversation_Manager SHALL allow session completion
6. WHEN conducting a check-in, THE Conversation_Manager SHALL limit the total conversation to a maximum of 10 minutes
7. WHEN the conversation exceeds 8 minutes, THE Eva_System SHALL begin wrapping up by asking only remaining critical questions

### Requirement 3: AI-Powered Data Extraction

**User Story:** As a system, I want to extract structured health data from conversational responses, so that health information can be stored, analyzed, and reported systematically.

#### Acceptance Criteria

1. WHEN a Check_In_Session is complete, THE Data_Extractor SHALL send the conversation history to Azure_OpenAI for processing
2. WHEN processing conversation history, THE Data_Extractor SHALL extract symptoms and pain descriptions as a list of strings
3. WHEN processing conversation history, THE Data_Extractor SHALL classify mood as positive, neutral, or negative
4. WHEN processing conversation history, THE Data_Extractor SHALL extract pain level as an integer between 0 and 10 if mentioned, or null if no pain reported
5. WHEN processing conversation history, THE Data_Extractor SHALL classify energy level as low, medium, or high
6. WHEN processing conversation history, THE Data_Extractor SHALL classify sleep quality as poor, fair, good, or excellent
7. WHEN processing conversation history, THE Data_Extractor SHALL determine medication adherence as yes, no, or partial
8. WHEN processing conversation history, THE Data_Extractor SHALL extract physical activities mentioned (sports, walks, exercise) as a list of strings
9. WHEN processing conversation history, THE Data_Extractor SHALL extract meal information (breakfast, lunch, dinner descriptions) as structured text
10. WHEN processing conversation history, THE Data_Extractor SHALL extract general feelings about the day as free text
11. WHEN processing conversation history, THE Data_Extractor SHALL prepare data structure for future fitness data integration from Health Connect
12. IF Azure_OpenAI extraction fails, THEN THE Eva_System SHALL store the raw transcript and mark the check-in for manual review

### Requirement 4: Medication Management

**User Story:** As a user, I want to track my medications and dosages, so that I can maintain an accurate medication list and monitor adherence.

#### Acceptance Criteria

1. WHEN a user adds a medication, THE Eva_System SHALL store the medication name, dosage, frequency, start date, and optional notes
2. WHEN a user requests their medication list, THE Eva_System SHALL return all active medications sorted by start date
3. WHEN a user updates a medication, THE Eva_System SHALL modify the specified medication record and preserve the medication ID
4. WHEN a user deletes a medication, THE Eva_System SHALL remove the medication record from the Health_Data_Store
5. WHEN a medication has an end date in the past, THE Eva_System SHALL mark it as inactive but retain the historical record
6. WHEN logging medication adherence, THE Eva_System SHALL create a MedicationLog entry with the medication ID and timestamp

### Requirement 5: Menstruation Cycle Tracking

**User Story:** As a user, I want to track my menstruation cycles, so that I can monitor patterns and share this information with my doctor.

#### Acceptance Criteria

1. WHEN a user logs menstruation data, THE Eva_System SHALL store the start date, optional end date, flow intensity, and symptoms
2. WHEN a user requests menstruation history, THE Eva_System SHALL return all cycles sorted by start date in descending order
3. WHEN storing flow intensity, THE Eva_System SHALL accept values: light, moderate, or heavy
4. WHEN storing symptoms, THE Eva_System SHALL accept a list of symptom strings
5. WHEN a new cycle is logged without an end date, THE Eva_System SHALL allow the record to remain open until updated

### Requirement 6: Blood Pressure Monitoring

**User Story:** As a user, I want to log blood pressure readings, so that I can track cardiovascular health over time.

#### Acceptance Criteria

1. WHEN a user logs a blood pressure reading, THE Eva_System SHALL store systolic value, diastolic value, pulse, and timestamp
2. WHEN storing systolic value, THE Eva_System SHALL validate that it is a positive integer between 70 and 250
3. WHEN storing diastolic value, THE Eva_System SHALL validate that it is a positive integer between 40 and 150
4. WHEN storing pulse, THE Eva_System SHALL validate that it is a positive integer between 30 and 220
5. WHEN a user requests blood pressure history, THE Eva_System SHALL return all readings sorted by timestamp in descending order

### Requirement 7: Dashboard Data Aggregation

**User Story:** As a user, I want to view aggregated health data on my dashboard, so that I can understand trends and patterns in my health.

#### Acceptance Criteria

1. WHEN a user requests dashboard summary, THE Eva_System SHALL return health metrics for the last 7 days by default
2. WHEN a user specifies a time range, THE Eva_System SHALL return data for the requested period (7, 30, or 90 days)
3. WHEN calculating trends, THE Eva_System SHALL compute average pain levels, mood distribution, and energy levels for the specified period
4. WHEN returning time-series data, THE Eva_System SHALL group data by date and include all available health metrics
5. WHEN no data exists for the requested period, THE Eva_System SHALL return an empty dataset with appropriate metadata

### Requirement 8: Doctor Report Generation

**User Story:** As a user, I want to generate comprehensive health reports in PDF format, so that I can share accurate health information with my doctor.

#### Acceptance Criteria

1. WHEN a user requests a report, THE Report_Generator SHALL create a PDF document containing all health data for the specified date range
2. WHEN generating a report, THE Eva_System SHALL include sections for: symptoms timeline, medication list, medication adherence, blood pressure trends, menstruation cycles, physical activities, meal patterns, and daily check-in summaries
3. WHEN a report is generated, THE Eva_System SHALL store the PDF file in Azure Blob Storage and create a Report record with file path
4. WHEN a user requests a report download, THE Eva_System SHALL return the PDF file from Azure Blob Storage
5. WHEN generating a report, THE Report_Generator SHALL format the document in a professional medical layout suitable for doctor consultation
6. WHEN no data exists for the specified date range, THE Eva_System SHALL generate a report indicating no data available for the period

### Requirement 9: API Documentation and Code Generation

**User Story:** As a mobile app developer, I want clear API documentation that drives code generation, so that the backend and mobile app stay in sync.

#### Acceptance Criteria

1. THE Eva_System SHALL define all API endpoints using OpenAPI specification in the /api folder
2. WHEN the OpenAPI specification is updated, THE Eva_System SHALL regenerate Go Gin endpoint functions using oapi-codegen
3. THE Eva_System SHALL use oapi-codegen with configuration file oapi-codegen-cfg.yaml to generate server code
4. WHEN generating code, THE Eva_System SHALL create type-safe request and response structures from the OpenAPI schema
5. THE Eva_System SHALL maintain the OpenAPI specification as the single source of truth for API contracts

### Requirement 10: Data Security and Privacy

**User Story:** As a user, I want my health data to be secure and private, so that my sensitive information is protected.

#### Acceptance Criteria

1. WHEN storing sensitive health data, THE Eva_System SHALL encrypt data at rest in the Health_Data_Store using AES-256 encryption
2. WHEN transmitting data, THE Eva_System SHALL require HTTPS for all API endpoints
3. WHEN a user requests data deletion, THE Eva_System SHALL remove all associated health records and mark the user account as deleted
4. WHEN a user requests data export, THE Eva_System SHALL provide all user data in machine-readable JSON format within 30 days
5. WHEN processing health data, THE Eva_System SHALL maintain an audit log of all data access and modifications

### Requirement 11: Request Validation and Error Handling

**User Story:** As a developer, I want comprehensive request validation and error handling, so that the API provides clear feedback and maintains data integrity.

#### Acceptance Criteria

1. WHEN an API request contains invalid data, THE Eva_System SHALL return a 400 Bad Request response with specific validation errors
2. WHEN an API request references a non-existent resource, THE Eva_System SHALL return a 404 Not Found response
3. WHEN an internal error occurs, THE Eva_System SHALL return a 500 Internal Server Error response and log the error details
4. WHEN validating request payloads, THE Eva_System SHALL check required fields, data types, and value ranges
5. WHEN an error response is returned, THE Eva_System SHALL include an error code, message, and optional details in JSON format
6. WHEN Azure_OpenAI or Azure_Speech_Service is unavailable, THE Eva_System SHALL return a 503 Service Unavailable response

### Requirement 12: Logging and Monitoring

**User Story:** As a system administrator, I want comprehensive logging and monitoring, so that I can troubleshoot issues and monitor system health.

#### Acceptance Criteria

1. WHEN an API request is received, THE Eva_System SHALL log the request method, path, user ID, and timestamp
2. WHEN an error occurs, THE Eva_System SHALL log the error message, stack trace, and request context
3. WHEN AI processing is performed, THE Eva_System SHALL log the processing time and token usage
4. WHEN a Check_In_Session is completed, THE Eva_System SHALL log the session duration and number of exchanges
5. WHEN database queries exceed 1 second, THE Eva_System SHALL log a slow query warning with the query details

### Requirement 13: API Versioning

**User Story:** As a mobile app developer, I want API versioning, so that the mobile app can handle backend changes gracefully.

#### Acceptance Criteria

1. THE Eva_System SHALL prefix all API endpoints with /api/v1/ to support versioning
2. WHEN the API version changes, THE Eva_System SHALL maintain backward compatibility for at least one previous version
3. WHEN an API endpoint is deprecated, THE Eva_System SHALL return a deprecation warning header for 90 days before removal
