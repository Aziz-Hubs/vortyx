-- =============================================================================
-- File: 001_vort.sql
-- Purpose: Database schema for VORT agent system
-- Created: 2026-02-15
-- =============================================================================
-- This SQL schema defines all tables required for the VORT agent system,
-- including agent registration, command execution, data collection, and monitoring.
-- Run this file against your PostgreSQL database to create the schema.
-- =============================================================================

-- =============================================================================
-- Table: agents
-- Purpose: Core agent registration and metadata storage
-- =============================================================================
-- Agents table stores all registered VORT agents with their authentication
-- credentials, system information, and operational status.
CREATE TABLE IF NOT EXISTS agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_key_hash VARCHAR(255) NOT NULL UNIQUE,    -- Hashed authentication key
    name VARCHAR(255) NOT NULL,                    -- Human-readable name
    hostname VARCHAR(255),                         -- Remote system hostname
    ip_address INET,                             -- Remote system IP
    os_type VARCHAR(50) NOT NULL,                -- OS type (linux, windows, darwin)
    os_version VARCHAR(100),                      -- OS version string
    arch VARCHAR(50),                             -- System architecture
    version VARCHAR(50) NOT NULL,                -- Agent software version
    capabilities JSONB DEFAULT '[]',              -- Supported features
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- Agent lifecycle status
    last_heartbeat TIMESTAMP WITH TIME ZONE,       -- Last communication
    registered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}',                   -- Custom metadata
    organization_id UUID,                         -- Associated organization
    tags TEXT[] DEFAULT '{}',                     -- Agent tags
    -- Ensure valid status values
    CHECK (status IN ('pending', 'active', 'inactive', 'disconnected', 'decommissioned'))
);

-- =============================================================================
-- Table: agent_groups
-- Purpose: Organizational grouping of agents
-- =============================================================================
-- Agent groups allow logical organization of agents for bulk operations
-- and access control.
CREATE TABLE IF NOT EXISTS agent_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    organization_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(name, organization_id)
);

-- =============================================================================
-- Table: agent_group_members
-- Purpose: Agent-group membership mapping
-- =============================================================================
CREATE TABLE IF NOT EXISTS agent_group_members (
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    group_id UUID REFERENCES agent_groups(id) ON DELETE CASCADE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (agent_id, group_id)
);

-- =============================================================================
-- Table: commands
-- Purpose: Command queue and execution tracking
-- =============================================================================
-- Commands table stores instructions to be executed by agents,
-- including scheduling, execution status, and results.
CREATE TABLE IF NOT EXISTS commands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    command_type VARCHAR(50) NOT NULL,           -- Command identifier (ping, shell, etc)
    payload JSONB NOT NULL,                     -- Command parameters
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- Execution state
    priority INTEGER DEFAULT 0,                 -- Execution priority
    scheduled_at TIMESTAMP WITH TIME ZONE,     -- Future execution time
    executed_at TIMESTAMP WITH TIME ZONE,      -- Start timestamp
    completed_at TIMESTAMP WITH TIME ZONE,     -- Completion timestamp
    result JSONB,                             -- Execution output
    error_message TEXT,                       -- Failure reason
    created_by UUID,                          -- Command initiator
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    -- Valid status values
    CHECK (status IN ('pending', 'dispatched', 'running', 'completed', 'failed', 'cancelled', 'timeout'))
);

-- =============================================================================
-- Table: command_history
-- Purpose: Audit trail of command executions
-- =============================================================================
CREATE TABLE IF NOT EXISTS command_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    command_id UUID REFERENCES commands(id) ON DELETE SET NULL,
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL,
    output JSONB,
    error_message TEXT,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =============================================================================
-- Table: agent_data
-- Purpose: Collected data from agents
-- =============================================================================
-- Agent data stores arbitrary data payloads collected by agents
-- during operation (metrics, inventory, etc).
CREATE TABLE IF NOT EXISTS agent_data (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    data_type VARCHAR(100) NOT NULL,           -- Data classification
    payload JSONB NOT NULL,                     -- Collected data
    collected_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    organization_id UUID,
    tags TEXT[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}'
);

-- =============================================================================
-- Table: agent_health
-- Purpose: System health metrics from agents
-- =============================================================================
-- Health metrics captured periodically from agents for monitoring.
CREATE TABLE IF NOT EXISTS agent_health (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    cpu_usage DECIMAL(5,2),                   -- CPU utilization %
    memory_usage DECIMAL(5,2),                -- Memory utilization %
    disk_usage DECIMAL(5,2),                 -- Disk utilization %
    network_in BIGINT,                        -- Bytes received
    network_out BIGINT,                       -- Bytes sent
    processes_count INTEGER,                  -- Running process count
    uptime_seconds BIGINT,                    -- System uptime
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =============================================================================
-- Table: agent_logs
-- Purpose: Centralized logging from agents
-- =============================================================================
-- Structured logs forwarded from agents for debugging and auditing.
CREATE TABLE IF NOT EXISTS agent_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    level VARCHAR(20) NOT NULL,               -- Log level
    message TEXT NOT NULL,                   -- Log content
    source VARCHAR(100),                     -- Component
    metadata JSONB DEFAULT '{}',
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =============================================================================
-- Table: agent_configs
-- Purpose: Agent runtime configuration
-- =============================================================================
-- Configuration values pushed to agents for behavior modification.
CREATE TABLE IF NOT EXISTS agent_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    config_key VARCHAR(255) NOT NULL,
    config_value JSONB NOT NULL,
    version INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(agent_id, config_key, is_active)
);

-- =============================================================================
-- Table: agent_secrets
-- Purpose: Encrypted secrets for agents
-- =============================================================================
-- Secure storage of encryption keys and secrets for agent communication.
CREATE TABLE IF NOT EXISTS agent_secrets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    secret_type VARCHAR(50) NOT NULL,
    encrypted_value BYTEA NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(agent_id, secret_type)
);

-- =============================================================================
-- Indexes
-- Purpose: Performance optimization for common queries
-- =============================================================================
CREATE INDEX idx_agents_status ON agents(status);
CREATE INDEX idx_agents_organization ON agents(organization_id);
CREATE INDEX idx_agents_last_heartbeat ON agents(last_heartbeat);
CREATE INDEX idx_commands_agent_status ON commands(agent_id, status);
CREATE INDEX idx_commands_scheduled ON commands(scheduled_at);
CREATE INDEX idx_agent_data_agent_collected ON agent_data(agent_id, collected_at);
CREATE INDEX idx_agent_health_agent_recorded ON agent_health(agent_id, recorded_at);
CREATE INDEX idx_agent_logs_agent_recorded ON agent_logs(agent_id, recorded_at);
CREATE INDEX idx_command_history_command ON command_history(command_id);

-- =============================================================================
-- Row Level Security
-- Purpose: Multi-tenant data isolation
-- =============================================================================
ALTER TABLE agents ENABLE ROW LEVEL SECURITY;
ALTER TABLE commands ENABLE ROW LEVEL SECURITY;
ALTER TABLE agent_data ENABLE ROW LEVEL SECURITY;
ALTER TABLE agent_health ENABLE ROW LEVEL SECURITY;
ALTER TABLE agent_logs ENABLE ROW LEVEL SECURITY;

-- =============================================================================
-- Triggers
-- Purpose: Automatic timestamp management
-- =============================================================================
-- Function: update_updated_at_column
-- Updates the updated_at timestamp on row modifications.
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger: agents
CREATE TRIGGER update_agents_updated_at
    BEFORE UPDATE ON agents
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger: commands
CREATE TRIGGER update_commands_updated_at
    BEFORE UPDATE ON commands
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger: agent_configs
CREATE TRIGGER update_agent_configs_updated_at
    BEFORE UPDATE ON agent_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger: agent_groups
CREATE TRIGGER update_agent_groups_updated_at
    BEFORE UPDATE ON agent_groups
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
