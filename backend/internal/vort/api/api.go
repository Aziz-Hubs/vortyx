// Package api provides HTTP handlers for VORT agent management.
// This package defines the HTTP API surface for registering, authenticating,
// and managing VORT agents. It handles request/response serialization and
// delegates business logic to the service layer.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	vortyxerrors "github.com/abdul/vortyx/backend/internal/pkg/errors"
	"github.com/abdul/vortyx/backend/internal/vort/auth"
	"github.com/abdul/vortyx/backend/internal/vort/db"
	"github.com/abdul/vortyx/backend/internal/vort/service"
)

// -----------------------------------------------------------------------------
// Handler Definitions
// -----------------------------------------------------------------------------

// AgentAPIHandler handles HTTP requests for VORT agent operations.
// It acts as a bridge between the HTTP layer and the service layer.
type AgentAPIHandler struct {
	agentService *service.AgentService    // Service layer for agent business logic.
	auth         *auth.AgentAuthenticator // Authenticator for agent credentials.
}

// NewAgentAPIHandler creates a new AgentAPIHandler with the provided dependencies.
func NewAgentAPIHandler(agentSvc *service.AgentService, auth *auth.AgentAuthenticator) *AgentAPIHandler {
	return &AgentAPIHandler{
		agentService: agentSvc,
		auth:         auth,
	}
}

// -----------------------------------------------------------------------------
// Request/Response DTOs
// -----------------------------------------------------------------------------

// RegisterAgentRequest defines the payload for registering a new VORT agent.
type RegisterAgentRequest struct {
	Name         string   `json:"name"`         // Human-readable name of the agent.
	Hostname     string   `json:"hostname"`     // Hostname of the machine running the agent.
	IPAddress    string   `json:"ip_address"`   // IP address of the agent.
	OSType       string   `json:"os_type"`      // Operating system type (e.g., "linux", "windows").
	OSVersion    string   `json:"os_version"`   // Operating system version.
	Arch         string   `json:"arch"`         // System architecture (e.g., "amd64", "arm64").
	Version      string   `json:"version"`      // Version of the VORT agent binary.
	Capabilities []string `json:"capabilities"` // List of capabilities supported by the agent.
	AgentKey     string   `json:"agent_key"`    // Secret key used for initial registration.
}

// RegisterAgentResponse defines the response after a successful agent registration.
type RegisterAgentResponse struct {
	AgentID  string                 `json:"agent_id"`         // Unique identifier for the registered agent.
	AgentKey string                 `json:"agent_key"`        // Secret key for the agent (returned once).
	Status   string                 `json:"status"`           // Current status of the agent (e.g., "pending", "active").
	Config   map[string]interface{} `json:"config,omitempty"` // Optional configuration for the agent.
}

// ErrorResponse defines a standard JSON error response format using the centralized error framework.
type ErrorResponse struct {
	Error   string `json:"error"`             // Short error message.
	Code    string `json:"code"`              // Machine-readable error code.
	Status  int    `json:"status"`            // HTTP status code.
	Details string `json:"details,omitempty"` // Additional error details.
}

// -----------------------------------------------------------------------------
// HTTP Handlers
// -----------------------------------------------------------------------------

// RegisterAgent handles POST requests to /api/v1/vort/agents/register.
// It registers a new VORT agent in the system.
func (h *AgentAPIHandler) RegisterAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeSimpleError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	resp, err := h.agentService.RegisterAgent(r.Context(), &service.AgentRegistrationRequest{
		Name:         req.Name,
		Hostname:     req.Hostname,
		IPAddress:    req.IPAddress,
		OSType:       req.OSType,
		OSVersion:    req.OSVersion,
		Arch:         req.Arch,
		Version:      req.Version,
		Capabilities: req.Capabilities,
		AgentKey:     req.AgentKey,
	})
	if err != nil {
		h.writeSimpleError(w, http.StatusInternalServerError, "failed to register agent", err.Error())
		return
	}

	h.writeJSON(w, http.StatusCreated, resp)
}

// AuthenticateAgent handles POST requests to /api/v1/vort/agents/authenticate.
// It authenticates a VORT agent and returns an access token.
func (h *AgentAPIHandler) AuthenticateAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type AuthRequest struct {
		AgentKey string `json:"agent_key"` // The agent's secret key.
		Secret   string `json:"secret"`    // Additional secret for authentication.
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeSimpleError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	token, err := h.agentService.AuthenticateAgent(r.Context(), req.AgentKey, req.Secret)
	if err != nil {
		h.writeSimpleError(w, http.StatusUnauthorized, "authentication failed", err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"token":      "token-placeholder", // TODO: Replace with actual JWT generation.
		"expires_at": token.ExpiresAt,
	})
}

// GetAgent handles GET requests to /api/v1/vort/agent.
// It retrieves details of a specific agent by ID (passed as query parameter "id").
func (h *AgentAPIHandler) GetAgent(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("id")
	if agentID == "" {
		h.writeSimpleError(w, http.StatusBadRequest, "missing agent id", "")
		return
	}

	agent, err := h.agentService.GetAgent(r.Context(), agentID)
	if err != nil {
		h.writeSimpleError(w, http.StatusNotFound, "agent not found", err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, agent)
}

// ListAgents handles GET requests to /api/v1/vort/agents.
// It lists agents with optional filtering by organization and status.
func (h *AgentAPIHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("organization_id")
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 50 // Default limit if not specified.
	}

	agents, err := h.agentService.ListAgents(r.Context(), strPtr(orgID), strPtr(status), limit, offset)
	if err != nil {
		h.writeSimpleError(w, http.StatusInternalServerError, "failed to list agents", err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"agents": agents,
		"limit":  limit,
		"offset": offset,
	})
}

// Heartbeat handles POST requests to /api/v1/vort/agent/heartbeat.
// Agents call this endpoint periodically to indicate they are still active.
func (h *AgentAPIHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		h.writeSimpleError(w, http.StatusBadRequest, "missing agent id", "")
		return
	}

	agent, err := h.agentService.Heartbeat(r.Context(), agentID)
	if err != nil {
		h.writeSimpleError(w, http.StatusInternalServerError, "failed to update heartbeat", err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, agent)
}

// SubmitData handles POST requests to /api/v1/vort/agent/data.
// Agents submit arbitrary telemetry data through this endpoint.
func (h *AgentAPIHandler) SubmitData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		h.writeSimpleError(w, http.StatusBadRequest, "missing agent id", "")
		return
	}

	type DataRequest struct {
		DataType string                 `json:"data_type"` // Type of data being submitted (e.g., "metrics", "inventory").
		Payload  map[string]interface{} `json:"payload"`   // The actual data payload.
	}

	var req DataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeSimpleError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	data, err := h.agentService.SubmitData(r.Context(), agentID, req.DataType, req.Payload)
	if err != nil {
		h.writeSimpleError(w, http.StatusInternalServerError, "failed to submit data", err.Error())
		return
	}

	h.writeJSON(w, http.StatusCreated, data)
}

// SubmitHealth handles POST requests to /api/v1/vort/agent/health.
// Agents submit their health status (CPU, memory, disk, etc.) through this endpoint.
func (h *AgentAPIHandler) SubmitHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		h.writeSimpleError(w, http.StatusBadRequest, "missing agent id", "")
		return
	}

	var health map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&health); err != nil {
		h.writeSimpleError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	healthData, err := h.agentService.SubmitHealth(r.Context(), agentID, health)
	if err != nil {
		h.writeSimpleError(w, http.StatusInternalServerError, "failed to submit health", err.Error())
		return
	}

	h.writeJSON(w, http.StatusCreated, healthData)
}

// SubmitLog handles POST requests to /api/v1/vort/agent/logs.
// Agents submit log entries through this endpoint for central logging.
func (h *AgentAPIHandler) SubmitLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		h.writeSimpleError(w, http.StatusBadRequest, "missing agent id", "")
		return
	}

	type LogRequest struct {
		Level    string                 `json:"level"`    // Log level (e.g., "info", "warn", "error").
		Message  string                 `json:"message"`  // Log message content.
		Metadata map[string]interface{} `json:"metadata"` // Additional context for the log.
	}

	var req LogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeSimpleError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if err := h.agentService.SubmitLog(r.Context(), agentID, req.Level, req.Message, req.Metadata); err != nil {
		h.writeSimpleError(w, http.StatusInternalServerError, "failed to submit log", err.Error())
		return
	}

	h.writeJSON(w, http.StatusAccepted, map[string]string{"status": "logged"})
}

// GetCommands handles GET requests to /api/v1/vort/agent/commands.
// Agents poll this endpoint to retrieve pending commands assigned to them.
func (h *AgentAPIHandler) GetCommands(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		h.writeSimpleError(w, http.StatusBadRequest, "missing agent id", "")
		return
	}

	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 50
	}

	commands, err := h.agentService.ListCommands(r.Context(), agentID, strPtr(status), limit, offset)
	if err != nil {
		h.writeSimpleError(w, http.StatusInternalServerError, "failed to list commands", err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"commands": commands,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetAgentStats handles GET requests to /api/v1/vort/stats.
// It returns aggregated statistics for agents within a specific organization.
func (h *AgentAPIHandler) GetAgentStats(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("organization_id")
	if orgID == "" {
		h.writeSimpleError(w, http.StatusBadRequest, "missing organization id", "")
		return
	}

	stats, err := h.agentService.GetAgentStats(r.Context(), orgID)
	if err != nil {
		h.writeSimpleError(w, http.StatusInternalServerError, "failed to get stats", err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, stats)
}

// -----------------------------------------------------------------------------
// Utility Functions
// -----------------------------------------------------------------------------

// writeJSON is a helper that writes a JSON response with the given status code.
func (h *AgentAPIHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Printf("Error encoding response: %v\n", err)
	}
}

// writeError writes a standardized JSON error response using the centralized error framework.
func (h *AgentAPIHandler) writeError(w http.ResponseWriter, err error) {
	var vortyxErr *vortyxerrors.VortyxError
	if errors.As(err, &vortyxErr) {
		status := vortyxErr.ToHTTPStatus()
		h.writeJSON(w, status, ErrorResponse{
			Error:   vortyxErr.Message,
			Code:    string(vortyxErr.Code),
			Status:  status,
			Details: vortyxErr.Details,
		})
		return
	}

	h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{
		Error:  "Internal Server Error",
		Code:   string(vortyxerrors.CodeInternal),
		Status: http.StatusInternalServerError,
	})
}

// writeSimpleError writes a simple error response without using the error framework.
func (h *AgentAPIHandler) writeSimpleError(w http.ResponseWriter, status int, message, details string) {
	h.writeJSON(w, status, ErrorResponse{
		Error:   message,
		Code:    "CUSTOM",
		Status:  status,
		Details: details,
	})
}

// strPtr converts a string to a pointer to a string.
// It's used to pass optional query parameters to the service layer.
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// -----------------------------------------------------------------------------
// Route Registration
// -----------------------------------------------------------------------------

// APIRoutes aggregates all VORT agent API handlers and provides route registration.
type APIRoutes struct {
	*AgentAPIHandler
}

// NewAPIRoutes creates a new APIRoutes instance with the required dependencies.
func NewAPIRoutes(agentSvc *service.AgentService, auth *auth.AgentAuthenticator) *APIRoutes {
	return &APIRoutes{
		AgentAPIHandler: NewAgentAPIHandler(agentSvc, auth),
	}
}

// RegisterRoutes registers all HTTP routes for the VORT agent API.
func (r *APIRoutes) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/vort/agents/register", r.RegisterAgent)
	mux.HandleFunc("/api/v1/vort/agents/authenticate", r.AuthenticateAgent)
	mux.HandleFunc("/api/v1/vort/agents", r.ListAgents)
	mux.HandleFunc("/api/v1/vort/agent", r.GetAgent)
	mux.HandleFunc("/api/v1/vort/agent/heartbeat", r.Heartbeat)
	mux.HandleFunc("/api/v1/vort/agent/data", r.SubmitData)
	mux.HandleFunc("/api/v1/vort/agent/health", r.SubmitHealth)
	mux.HandleFunc("/api/v1/vort/agent/logs", r.SubmitLog)
	mux.HandleFunc("/api/v1/vort/agent/commands", r.GetCommands)
	mux.HandleFunc("/api/v1/vort/stats", r.GetAgentStats)
}

// -----------------------------------------------------------------------------
// Response Transformers
// -----------------------------------------------------------------------------

// AgentResponse defines the JSON structure for agent details returned to clients.
type AgentResponse struct {
	ID             string                 `json:"id"`              // Unique identifier.
	Name           string                 `json:"name"`            // Human-readable name.
	Hostname       *string                `json:"hostname"`        // Machine hostname.
	IPAddress      *string                `json:"ip_address"`      // Agent IP address.
	OSType         string                 `json:"os_type"`         // Operating system type.
	OSVersion      *string                `json:"os_version"`      // OS version string.
	Arch           *string                `json:"arch"`            // System architecture.
	Version        string                 `json:"version"`         // Agent version.
	Capabilities   []string               `json:"capabilities"`    // Supported features.
	Status         string                 `json:"status"`          // Current operational status.
	LastHeartbeat  *time.Time             `json:"last_heartbeat"`  // Time of last heartbeat.
	RegisteredAt   time.Time              `json:"registered_at"`   // Registration timestamp.
	UpdatedAt      time.Time              `json:"updated_at"`      // Last update timestamp.
	Metadata       map[string]interface{} `json:"metadata"`        // Custom metadata.
	OrganizationID *string                `json:"organization_id"` // Associated organization.
	Tags           []string               `json:"tags"`            // Agent tags.
}

// convertAgentToResponse transforms a database Agent model into an API AgentResponse.
// This separation allows the API layer to evolve independently from the database schema.
func convertAgentToResponse(agent *db.Agent) *AgentResponse {
	resp := &AgentResponse{
		ID:      agent.ID.String(),
		Name:    agent.Name,
		Version: agent.Version,
		Status:  agent.Status,
	}

	if agent.Hostname.Valid {
		resp.Hostname = &agent.Hostname.String
	}
	if agent.IpAddress != nil {
		ip := agent.IpAddress.String()
		resp.IPAddress = &ip
	}
	resp.OSType = agent.OsType
	if agent.OsVersion.Valid {
		resp.OSVersion = &agent.OsVersion.String
	}
	if agent.Arch.Valid {
		resp.Arch = &agent.Arch.String
	}
	if agent.Capabilities != nil {
		var caps []string
		json.Unmarshal(agent.Capabilities, &caps)
		resp.Capabilities = caps
	}
	if agent.LastHeartbeat.Valid {
		resp.LastHeartbeat = &agent.LastHeartbeat.Time
	}
	if agent.RegisteredAt.Valid {
		resp.RegisteredAt = agent.RegisteredAt.Time
	}
	if agent.UpdatedAt.Valid {
		resp.UpdatedAt = agent.UpdatedAt.Time
	}
	if agent.Metadata != nil {
		var meta map[string]interface{}
		json.Unmarshal(agent.Metadata, &meta)
		resp.Metadata = meta
	}
	if agent.OrganizationID.Valid {
		orgID := agent.OrganizationID.String()
		resp.OrganizationID = &orgID
	}
	resp.Tags = agent.Tags

	return resp
}
