-- Add check-in related tables for Eva Health Assistant

CREATE TABLE IF NOT EXISTS check_in_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP,
    expired_at TIMESTAMP,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS conversation_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES check_in_sessions(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    audio_file_path VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS audio_recordings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES check_in_sessions(id) ON DELETE CASCADE,
    message_id UUID REFERENCES conversation_messages(id) ON DELETE SET NULL,
    file_path VARCHAR(500) NOT NULL,
    duration_seconds FLOAT,
    transcription TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS health_check_ins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    session_id UUID REFERENCES check_in_sessions(id) ON DELETE SET NULL,
    check_in_date DATE NOT NULL,
    symptoms TEXT[],
    mood VARCHAR(50),
    pain_level INTEGER CHECK (pain_level >= 0 AND pain_level <= 10),
    energy_level VARCHAR(50),
    sleep_quality VARCHAR(50),
    medication_taken VARCHAR(50),
    physical_activity TEXT[],
    breakfast TEXT,
    lunch TEXT,
    dinner TEXT,
    general_feeling TEXT,
    additional_notes TEXT,
    raw_transcript TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS medications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    dosage VARCHAR(255) NOT NULL,
    frequency VARCHAR(255) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    notes TEXT,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS medication_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    medication_id UUID NOT NULL REFERENCES medications(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    taken_at TIMESTAMP NOT NULL DEFAULT NOW(),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS menstruation_cycles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    flow_intensity VARCHAR(50),
    symptoms TEXT[],
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS blood_pressure_readings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    systolic INTEGER NOT NULL CHECK (systolic >= 70 AND systolic <= 250),
    diastolic INTEGER NOT NULL CHECK (diastolic >= 40 AND diastolic <= 150),
    pulse INTEGER CHECK (pulse >= 30 AND pulse <= 220),
    measured_at TIMESTAMP NOT NULL DEFAULT NOW(),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS fitness_data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    date DATE NOT NULL,
    data_type VARCHAR(50) NOT NULL,
    value FLOAT NOT NULL,
    unit VARCHAR(50) NOT NULL,
    source VARCHAR(100),
    source_data_id VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_check_in_sessions_user_id ON check_in_sessions(user_id);
CREATE INDEX idx_check_in_sessions_status ON check_in_sessions(status);
CREATE INDEX idx_conversation_messages_session_id ON conversation_messages(session_id);
CREATE INDEX idx_audio_recordings_session_id ON audio_recordings(session_id);
CREATE INDEX idx_audio_recordings_message_id ON audio_recordings(message_id);
CREATE INDEX idx_health_check_ins_user_id ON health_check_ins(user_id);
CREATE INDEX idx_health_check_ins_date ON health_check_ins(check_in_date);
CREATE INDEX idx_medications_user_id ON medications(user_id);
CREATE INDEX idx_medications_active ON medications(active);
CREATE INDEX idx_medication_logs_medication_id ON medication_logs(medication_id);
CREATE INDEX idx_medication_logs_user_id ON medication_logs(user_id);
CREATE INDEX idx_menstruation_cycles_user_id ON menstruation_cycles(user_id);
CREATE INDEX idx_menstruation_cycles_start_date ON menstruation_cycles(start_date);
CREATE INDEX idx_blood_pressure_readings_user_id ON blood_pressure_readings(user_id);
CREATE INDEX idx_blood_pressure_readings_measured_at ON blood_pressure_readings(measured_at);
CREATE INDEX idx_fitness_data_user_id ON fitness_data(user_id);
CREATE INDEX idx_fitness_data_date ON fitness_data(date);
CREATE INDEX idx_fitness_data_source_data_id ON fitness_data(source_data_id);
CREATE INDEX idx_reports_user_id ON reports(user_id);
CREATE INDEX idx_reports_status ON reports(status);
