-- Create audit_logs table for GDPR compliance
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    operation_type VARCHAR(50) NOT NULL, -- CREATE, UPDATE, DELETE, READ
    resource_type VARCHAR(100) NOT NULL, -- health_check_in, medication, etc.
    resource_id VARCHAR(255) NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    ip_address VARCHAR(45), -- IPv4 or IPv6
    user_agent TEXT,
    additional_data JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for efficient querying
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);
CREATE INDEX idx_audit_logs_resource_type ON audit_logs(resource_type);
CREATE INDEX idx_audit_logs_operation_type ON audit_logs(operation_type);
