-- name: CreateAuditLog :one
INSERT INTO audit_logs (
    user_id, username, action, resource_type, resource_id, details
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: ListAuditLogs :many
SELECT * FROM audit_logs
WHERE 
    (sqlc.narg('user_id')::text IS NULL OR user_id = sqlc.narg('user_id'))
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountAuditLogs :one
SELECT COUNT(*) FROM audit_logs
WHERE (sqlc.narg('user_id')::text IS NULL OR user_id = sqlc.narg('user_id'));

-- name: GetSystemConfig :one
SELECT * FROM system_config
WHERE key = $1;

-- name: SetSystemConfig :one
INSERT INTO system_config (key, value, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (key) DO UPDATE
SET value = $2, updated_at = NOW()
RETURNING *;
