-- =============================================================================
-- File: query.sql
-- Purpose: SQL queries for VORT agent system database operations
-- Created: 2026-02-15
-- =============================================================================
-- This file contains named SQL queries for all VORT agent system operations.
-- These queries are used with sqlc or similar tools to generate type-safe Go code.
-- =============================================================================

-- =============================================================================
-- Agent Queries
-- =============================================================================

-- name: GetAgentByID :one
-- Retrieve a single agent by UUID
SELECT * FROM agents WHERE id = $1;

-- name: GetAgentByKeyHash :one
-- Retrieve agent by authentication key hash
SELECT * FROM agents WHERE agent_key_hash = $1;

-- name: ListAgents :many
-- List agents with optional filtering
SELECT * FROM agents
WHERE ($1::uuid IS NULL OR organization_id = $1)
  AND ($2::varchar IS NULL OR status = $2)
ORDER BY registered_at DESC
LIMIT $3 OFFSET $4;

-- name: CreateAgent :one
-- Register a new agent
INSERT INTO agents (
    agent_key_hash, name, hostname, ip_address, os_type, os_version,
    arch, version, capabilities, status, organization_id, tags, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: UpdateAgentStatus :one
-- Update agent status
UPDATE agents
SET status = $2, last_heartbeat = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateAgentHeartbeat :one
-- Record agent heartbeat
UPDATE agents
SET last_heartbeat = NOW(), status = 'active', updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteAgent :exec
-- Permanently remove agent
DELETE FROM agents WHERE id = $1;

-- =============================================================================
-- Command Queries
-- =============================================================================

-- name: CreateCommand :one
-- Queue a new command for execution
INSERT INTO commands (
    agent_id, command_type, payload, status, priority,
    scheduled_at, created_by
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetCommand :one
-- Retrieve command by ID
SELECT * FROM commands WHERE id = $1;

-- name: GetPendingCommands :many
-- Fetch pending commands for an agent
SELECT * FROM commands
WHERE agent_id = $1
  AND status IN ('pending', 'dispatched')
  AND (scheduled_at IS NULL OR scheduled_at <= NOW())
ORDER BY priority DESC, created_at ASC
LIMIT $2;

-- name: UpdateCommandStatus :one
-- Update command execution status
UPDATE commands
SET status = $2,
    executed_at = CASE WHEN $2 = 'running' THEN NOW() ELSE executed_at END,
    completed_at = CASE WHEN $2 IN ('completed', 'failed', 'cancelled', 'timeout') THEN NOW() ELSE completed_at END,
    result = COALESCE($3, result),
    error_message = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListCommands :many
-- List commands for an agent
SELECT * FROM commands
WHERE agent_id = $1
  AND ($2::varchar IS NULL OR status = $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CreateCommandHistory :exec
-- Record command history entry
INSERT INTO command_history (command_id, agent_id, status, output, error_message)
VALUES ($1, $2, $3, $4, $5);

-- =============================================================================
-- Agent Data Queries
-- =============================================================================

-- name: CreateAgentData :one
-- Store collected data from agent
INSERT INTO agent_data (agent_id, data_type, payload, organization_id, tags, metadata)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListAgentData :many
-- Query collected agent data
SELECT * FROM agent_data
WHERE agent_id = $1
  AND ($2::varchar IS NULL OR data_type = $2)
ORDER BY collected_at DESC
LIMIT $3 OFFSET $4;

-- =============================================================================
-- Health Monitoring Queries
-- =============================================================================

-- name: CreateAgentHealth :one
-- Store health metrics
INSERT INTO agent_health (
    agent_id, cpu_usage, memory_usage, disk_usage,
    network_in, network_out, processes_count, uptime_seconds
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetLatestHealth :one
-- Get most recent health data
SELECT * FROM agent_health
WHERE agent_id = $1
ORDER BY recorded_at DESC
LIMIT 1;

-- name: ListAgentHealth :many
-- Get health history
SELECT * FROM agent_health
WHERE agent_id = $1
ORDER BY recorded_at DESC
LIMIT $2;

-- =============================================================================
-- Logging Queries
-- =============================================================================

-- name: CreateAgentLog :exec
-- Store log entry from agent
INSERT INTO agent_logs (agent_id, level, message, source, metadata)
VALUES ($1, $2, $3, $4, $5);

-- name: ListAgentLogs :many
-- Query agent logs
SELECT * FROM agent_logs
WHERE agent_id = $1
  AND ($2::varchar IS NULL OR level = $2)
ORDER BY recorded_at DESC
LIMIT $3 OFFSET $4;

-- =============================================================================
-- Configuration Queries
-- =============================================================================

-- name: CreateAgentConfig :one
-- Create or update agent configuration
INSERT INTO agent_configs (agent_id, config_key, config_value, is_active)
VALUES ($1, $2, $3, $4)
ON CONFLICT (agent_id, config_key, is_active)
DO UPDATE SET config_value = $3, version = agent_configs.version + 1, updated_at = NOW()
RETURNING *;

-- name: GetAgentConfig :one
-- Retrieve active configuration value
SELECT * FROM agent_configs
WHERE agent_id = $1 AND config_key = $2 AND is_active = true;

-- name: ListAgentConfigs :many
-- Get all active configurations
SELECT * FROM agent_configs
WHERE agent_id = $1 AND is_active = true
ORDER BY config_key;

-- =============================================================================
-- Group Management Queries
-- =============================================================================

-- name: CreateAgentGroup :one
-- Create new agent group
INSERT INTO agent_groups (name, description, organization_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: AddAgentToGroup :exec
-- Add agent to group
INSERT INTO agent_group_members (agent_id, group_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveAgentFromGroup :exec
-- Remove agent from group
DELETE FROM agent_group_members WHERE agent_id = $1 AND group_id = $2;

-- name: GetAgentsByGroup :many
-- List agents in a group
SELECT a.* FROM agents a
JOIN agent_group_members agm ON a.id = agm.agent_id
WHERE agm.group_id = $1;

-- name: GetAgentGroups :many
-- Get groups for an agent
SELECT ag.* FROM agent_groups ag
JOIN agent_group_members agm ON ag.id = agm.group_id
WHERE agm.agent_id = $1;

-- =============================================================================
-- Secrets Management Queries
-- =============================================================================

-- name: CreateAgentSecret :one
-- Store encrypted secret
INSERT INTO agent_secrets (agent_id, secret_type, encrypted_value, expires_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (agent_id, secret_type)
DO UPDATE SET encrypted_value = $3, expires_at = $4
RETURNING *;

-- name: GetAgentSecret :one
-- Retrieve active secret
SELECT * FROM agent_secrets
WHERE agent_id = $1 AND secret_type = $2
  AND (expires_at IS NULL OR expires_at > NOW());

-- =============================================================================
-- Statistics Queries
-- =============================================================================

-- name: GetAgentStats :one
-- Get agent count by status for organization
SELECT
    COUNT(*) FILTER (WHERE status = 'active') as active_count,
    COUNT(*) FILTER (WHERE status = 'inactive') as inactive_count,
    COUNT(*) FILTER (WHERE status = 'disconnected') as disconnected_count,
    COUNT(*) as total_count
FROM agents
WHERE organization_id = $1;
