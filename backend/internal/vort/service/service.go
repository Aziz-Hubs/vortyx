// =============================================================================
// Package: service
// File: service.go
// Purpose: Core agent service implementation for VORT system
// Created: 2026-02-15
// =============================================================================
// This package provides the core service layer for VORT agent management,
// including command execution, data collection, and agent lifecycle management.
//
// Authentication: This service uses Zitadel for all authentication.
// - Human users authenticate via OIDC (Zitadel)
// - Machine agents (VORT) authenticate via Zitadel Machine Users
// - Agent registration and authentication endpoints are public
// - All other endpoints require valid Zitadel JWT tokens
// =============================================================================

package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/netip"
	"sync"
	"time"

	"connectrpc.com/connect"
	vortv1 "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/vort/v1"
	"github.com/abdul/vortyx/backend/gen/proto/go/vortyx/vort/v1/vortv1connect"
	"github.com/abdul/vortyx/backend/internal/pkg/common"
	"github.com/abdul/vortyx/backend/internal/vort/config"
	"github.com/abdul/vortyx/backend/internal/vort/db"
	"github.com/abdul/vortyx/backend/internal/vort/encryption"
	"github.com/abdul/vortyx/backend/internal/vort/machineuser"
	"github.com/abdul/vortyx/backend/internal/vort/mq"
	"github.com/abdul/vortyx/backend/internal/vort/token"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// =============================================================================
// Type: AgentService
// Purpose: Main service for managing VORT agents
// =============================================================================
// AgentService provides the core business logic for agent management,
// including registration, command dispatch, data collection, and health monitoring.
//
// Thread Safety:
//   - Uses sync.RWMutex for concurrent access
//   - Background goroutines for command dispatch and health monitoring
//
// Authentication Note:
//   - All authentication is handled by Zitadel via interceptors
//   - This service focuses on business logic, not authentication
type AgentService struct {
	vortv1connect.UnimplementedVortServiceHandler // Embed for forward compatibility

	mu            sync.RWMutex
	queries       db.Querier                // Database interface
	config        *config.Config            // Configuration
	encryptor     *encryption.KeyManager    // Encryption key manager
	tokenSvc      *token.AgentTokenService  // JWT token service
	machineUserAuth *machineuser.MachineUserAuth // Zitadel machine user auth
	publisher     mq.Publisher              // Message queue publisher
	subscribers   map[string]mq.Subscriber  // Message subscribers
	handlers      map[string]CommandHandler // Command handlers
	shutdownCh    chan struct{}             // Shutdown signal
	wg            sync.WaitGroup            // Wait group for goroutines
}

// =============================================================================
// Type: CommandHandler
// Purpose: Function type for handling commands
// =============================================================================
// CommandHandler processes commands received by agents.
// Returns command result or error.
type CommandHandler func(ctx context.Context, cmd *db.Command) (map[string]interface{}, error)

// =============================================================================
// Constructor
// =============================================================================

// NewService creates a new AgentService instance.
// Authentication is handled by Zitadel interceptors, not by this service.
func NewService(
	queries db.Querier,
	cfg *config.Config,
	encryptor *encryption.KeyManager,
) *AgentService {
	tokenSvc, err := token.GetAgentTokenService()
	if err != nil {
		cfg.Logger.Error().Err(err).Msg("failed to initialize agent token service")
	}

	machineUserAuth, err := machineuser.GetMachineUserAuth(context.Background())
	if err != nil {
		cfg.Logger.Warn().Err(err).Msg("Zitadel machine user auth not configured - using fallback authentication")
	}

	svc := &AgentService{
		queries:         queries,
		config:          cfg,
		encryptor:       encryptor,
		tokenSvc:        tokenSvc,
		machineUserAuth: machineUserAuth,
		subscribers:     make(map[string]mq.Subscriber),
		handlers:        make(map[string]CommandHandler),
		shutdownCh:      make(chan struct{}),
	}

	// Register default command handlers
	svc.registerDefaultHandlers()

	return svc
}

// NewServiceFromPool creates a new AgentService using a database connection pool.
// This helper function initializes all dependencies with default configurations.
//
// Authentication: Uses Zitadel for all authentication via interceptors.
// No custom JWT implementation is needed - Zitadel handles token validation.
//
// Parameters:
//   - pool: Database connection pool
//
// Returns:
//   - *AgentService: Configured service
func NewServiceFromPool(pool *pgxpool.Pool) *AgentService {
	// Initialize dependencies
	queries := db.New(pool)

	cfg, err := config.Load()
	if err != nil {
		// Fallback to default config if load fails
		cfg = &config.Config{}
		// In a real app we might want to panic or log error, but for now safe fallback
	}

	// Encryption key manager for secure agent communication
	encryptor := encryption.NewKeyManager(30) // 30-day rotation

	return NewService(queries, cfg, encryptor)
}

// =============================================================================
// Service Lifecycle
// =============================================================================

// SetPublisher configures the message queue publisher.
func (s *AgentService) SetPublisher(p mq.Publisher) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.publisher = p
}

// Start initializes the service and starts background workers.
func (s *AgentService) Start(ctx context.Context) error {
	return nil
}

// Stop gracefully shuts down the service.
func (s *AgentService) Stop(ctx context.Context) error {
	close(s.shutdownCh)
	return nil
}

// =============================================================================
// Agent Management (ConnectRPC Implementation)
// =============================================================================
// Note: Authentication is handled by Zitadel interceptors.
// RegisterAgent and AuthenticateAgent are public endpoints (no auth required).
// All other endpoints require valid Zitadel JWT tokens.

// RegisterAgent registers a new agent with the system.
func (s *AgentService) RegisterAgent(ctx context.Context, req *connect.Request[vortv1.RegisterAgentRequest]) (*connect.Response[vortv1.RegisterAgentResponse], error) {
	msg := req.Msg

	// Validation
	if err := common.ValidateAgentName(msg.Name); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if msg.Hostname != "" {
		if err := common.ValidateHostname(msg.Hostname); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
	}
	if msg.IpAddress != "" {
		if err := common.ValidateIP(msg.IpAddress); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
	}
	if msg.OsType != "" {
		if err := common.ValidateOSType(msg.OsType); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
	}

	// Hash the agent key for secure storage (simple SHA256 for now)
	agentKeyHash := simpleHash(msg.AgentKey)

	// Prepare capabilities as JSON
	capabilitiesJSON, err := json.Marshal(msg.Capabilities)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to marshal capabilities: %w", err))
	}

	// Parse IP address
	var ipAddr *netip.Addr
	if msg.IpAddress != "" {
		addr, err := netip.ParseAddr(msg.IpAddress)
		if err == nil {
			ipAddr = &addr
		}
	}

	// Create the agent in the database
	agent, err := s.queries.CreateAgent(ctx, db.CreateAgentParams{
		AgentKeyHash:   agentKeyHash,
		Name:           msg.Name,
		Hostname:       pgtype.Text{String: msg.Hostname, Valid: msg.Hostname != ""},
		IpAddress:      ipAddr,
		OsType:         msg.OsType,
		OsVersion:      pgtype.Text{String: msg.OsVersion, Valid: msg.OsVersion != ""},
		Arch:           pgtype.Text{String: msg.Arch, Valid: msg.Arch != ""},
		Version:        msg.Version,
		Capabilities:   capabilitiesJSON,
		Status:         "pending",
		OrganizationID: pgtype.UUID{Valid: false}, // TODO: Handle Org ID from request or token
		Tags:           msg.Tags,
		Metadata:       []byte("{}"),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create agent: %w", err))
	}

	// Generate a secure agent key for the agent to use
	generatedKey := generateSecureKey()

	// Hash the secret with bcrypt before storing
	hashedSecret, err := hashPassword(generatedKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to hash agent secret: %w", err))
	}

	// Store the agent secret
	_, err = s.queries.CreateAgentSecret(ctx, db.CreateAgentSecretParams{
		AgentID:        agent.ID,
		SecretType:     "agent_key",
		EncryptedValue: []byte(hashedSecret),
		ExpiresAt:      pgtype.Timestamptz{Time: time.Now().Add(365 * 24 * time.Hour), Valid: true},
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create agent secret: %w", err))
	}

	// Create config struct
	configStruct, _ := structpb.NewStruct(map[string]interface{}{
		"api_version": "v1",
		"server":      s.config.GetServerAddress(),
		"heartbeat":   s.config.GetAgentHeartbeatInterval().String(),
	})

	// Return the registration response
	return connect.NewResponse(&vortv1.RegisterAgentResponse{
		AgentId:  agent.ID.String(),
		AgentKey: generatedKey,
		Status:   agent.Status,
		Config:   configStruct,
	}), nil
}

// AuthenticateAgent authenticates an agent using Zitadel Machine User flow.
// This endpoint validates the agent's credentials and returns a Zitadel JWT token
// using JWT Profile Grant (urn:ietf:params:oauth:grant-type:jwt-bearer).
//
// Authentication Flow:
// 1. Validate agent key and secret against database
// 2. Verify agent status is active/pending
// 3. If Zitadel machine user auth is configured, use JWT Profile Grant to obtain a Zitadel token
// 4. Return the Zitadel JWT token for subsequent API calls
//
// The machine user authentication uses:
// - VORT_MACHINE_USER_KEY_PATH or VORT_MACHINE_USER_KEY: RSA private key for JWT signing
// - VORT_MACHINE_USER_KEY_ID: Key ID for the JWT header
// - ZITADEL_ISSUER: Zitadel instance URL
func (s *AgentService) AuthenticateAgent(ctx context.Context, req *connect.Request[vortv1.AuthenticateAgentRequest]) (*connect.Response[vortv1.AuthenticateAgentResponse], error) {
	msg := req.Msg

	if err := common.ValidateRequired(msg.AgentKey); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("agent key is required"))
	}
	if err := common.ValidateRequired(msg.Secret); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("secret is required"))
	}

	// 1. Hash the provided Agent Key
	keyHash := simpleHash(msg.AgentKey)

	// 2. Find Agent by Key Hash
	agent, err := s.queries.GetAgentByKeyHash(ctx, keyHash)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid credentials"))
	}

	// 3. Get Agent Secret
	secret, err := s.queries.GetAgentSecret(ctx, db.GetAgentSecretParams{
		AgentID:    agent.ID,
		SecretType: "agent_key",
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid credentials"))
	}

	// 4. Verify Secret with bcrypt
	storedHash := string(secret.EncryptedValue)

	if !checkPassword(storedHash, msg.Secret) {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid credentials"))
	}

	// 5. Check Agent Status
	if agent.Status != "active" && agent.Status != "pending" {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("agent is %s", agent.Status))
	}

	// 6. Generate JWT token using Zitadel Machine User flow
	var jwtToken string
	var expiresAt time.Time
	tokenExpiry := 24 * time.Hour

	// Try Zitadel machine user authentication first
	if s.machineUserAuth != nil {
		zitadelToken, exp, err := s.machineUserAuth.IssueToken(agent.ID.String())
		if err != nil {
			s.config.Logger.Error().Err(err).Str("agent_id", agent.ID.String()).Msg("failed to get Zitadel token, falling back to internal token")
		} else {
			jwtToken = zitadelToken
			expiresAt = exp
		}
	}

	// Fallback to internal token service if Zitadel failed or is not configured
	if jwtToken == "" && s.tokenSvc != nil {
		generatedToken, err := s.tokenSvc.IssueToken(agent.ID.String(), agent.Name, tokenExpiry)
		if err != nil {
			s.config.Logger.Error().Err(err).Str("agent_id", agent.ID.String()).Msg("failed to generate fallback JWT token")
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to generate authentication token"))
		}
		jwtToken = generatedToken
		expiresAt = time.Now().Add(tokenExpiry)
	} else if jwtToken == "" {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("no token service available"))
	}

	return connect.NewResponse(&vortv1.AuthenticateAgentResponse{
		Token:     jwtToken,
		ExpiresAt: timestamppb.New(expiresAt),
	}), nil
}

// Heartbeat processes agent heartbeat.
func (s *AgentService) Heartbeat(ctx context.Context, req *connect.Request[vortv1.HeartbeatRequest]) (*connect.Response[vortv1.HeartbeatResponse], error) {
	msg := req.Msg
	if err := common.ValidateUUID(msg.AgentId); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid agent id: %w", err))
	}

	agentUUID, err := uuid.Parse(msg.AgentId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid agent id: %w", err))
	}

	// Update agent status and heartbeat
	// Use UpdateAgentStatus if status is provided, otherwise just heartbeat
	var agent db.Agent
	if msg.Status != "" {
		agent, err = s.queries.UpdateAgentStatus(ctx, db.UpdateAgentStatusParams{
			ID:     pgtype.UUID{Bytes: agentUUID, Valid: true},
			Status: msg.Status,
		})
	} else {
		agent, err = s.queries.UpdateAgentHeartbeat(ctx, pgtype.UUID{Bytes: agentUUID, Valid: true})
	}

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update heartbeat: %w", err))
	}

	// Check for pending commands
	pendingCommands, err := s.queries.GetPendingCommands(ctx, db.GetPendingCommandsParams{
		AgentID: pgtype.UUID{Bytes: agentUUID, Valid: true},
		Limit:   1,
	})
	commandPending := err == nil && len(pendingCommands) > 0

	return connect.NewResponse(&vortv1.HeartbeatResponse{
		Status:         agent.Status,
		CommandPending: commandPending,
	}), nil
}

// GetAgent retrieves agent details.
func (s *AgentService) GetAgent(ctx context.Context, req *connect.Request[vortv1.GetAgentRequest]) (*connect.Response[vortv1.GetAgentResponse], error) {
	msg := req.Msg
	if err := common.ValidateUUID(msg.AgentId); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid agent id: %w", err))
	}

	agentUUID, err := uuid.Parse(msg.AgentId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid agent id: %w", err))
	}

	agent, err := s.queries.GetAgentByID(ctx, pgtype.UUID{Bytes: agentUUID, Valid: true})
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("agent not found: %w", err))
	}

	return connect.NewResponse(&vortv1.GetAgentResponse{
		Agent: convertToProtoAgent(agent),
	}), nil
}

// ListAgents lists agents.
func (s *AgentService) ListAgents(ctx context.Context, req *connect.Request[vortv1.ListAgentsRequest]) (*connect.Response[vortv1.ListAgentsResponse], error) {
	msg := req.Msg

	limit := int32(10)
	if msg.Limit > 0 {
		limit = msg.Limit
	}

	var orgID pgtype.UUID
	if msg.OrganizationId != "" {
		uid, err := uuid.Parse(msg.OrganizationId)
		if err == nil {
			orgID = pgtype.UUID{Bytes: uid, Valid: true}
		}
	}

	agents, err := s.queries.ListAgents(ctx, db.ListAgentsParams{
		Column1: orgID,
		Column2: msg.Status,
		Limit:   limit,
		Offset:  msg.Offset,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list agents: %w", err))
	}

	protoAgents := make([]*vortv1.Agent, len(agents))
	for i, a := range agents {
		protoAgents[i] = convertToProtoAgent(a)
	}

	return connect.NewResponse(&vortv1.ListAgentsResponse{
		Agents:     protoAgents,
		TotalCount: int32(len(agents)), // TODO: Implement Count query for proper pagination
	}), nil
}

// SubmitData submits agent data.
func (s *AgentService) SubmitData(ctx context.Context, req *connect.Request[vortv1.SubmitDataRequest]) (*connect.Response[vortv1.SubmitDataResponse], error) {
	msg := req.Msg
	if err := common.ValidateUUID(msg.AgentId); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid agent id: %w", err))
	}
	if err := common.ValidateRequired(msg.DataType); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("data type is required"))
	}

	agentUUID, err := uuid.Parse(msg.AgentId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid agent id: %w", err))
	}

	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid payload: %w", err))
	}

	data, err := s.queries.CreateAgentData(ctx, db.CreateAgentDataParams{
		AgentID:  pgtype.UUID{Bytes: agentUUID, Valid: true},
		DataType: msg.DataType,
		Payload:  payloadBytes,
		Metadata: []byte("{}"),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to submit data: %w", err))
	}

	return connect.NewResponse(&vortv1.SubmitDataResponse{
		DataId: data.ID.String(),
		Status: "received",
	}), nil
}

// =============================================================================
// Helper Functions
// =============================================================================

func convertToProtoAgent(a db.Agent) *vortv1.Agent {
	var caps []string
	if len(a.Capabilities) > 0 {
		_ = json.Unmarshal(a.Capabilities, &caps)
	}

	return &vortv1.Agent{
		Id:            a.ID.String(),
		Name:          a.Name,
		Hostname:      a.Hostname.String,
		IpAddress:     a.IpAddress.String(),
		OsType:        a.OsType,
		OsVersion:     a.OsVersion.String,
		Arch:          a.Arch.String,
		Version:       a.Version,
		Status:        a.Status,
		LastHeartbeat: timestamppb.New(a.LastHeartbeat.Time),
		RegisteredAt:  timestamppb.New(a.RegisteredAt.Time),
		UpdatedAt:     timestamppb.New(a.UpdatedAt.Time),
		Tags:          a.Tags,
		Capabilities:  caps,
	}
}

func hashPassword(s string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func checkPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// simpleHash creates a simple SHA256 hash of a string.
// Used for agent keys where we need deterministic hashing for lookups.
func simpleHash(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}

// generateSecureKey generates a random secure key.
func generateSecureKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// =============================================================================
// Command Handler Registration
// =============================================================================

// RegisterCommandHandler registers a handler for a command type.
func (s *AgentService) RegisterCommandHandler(commandType string, handler CommandHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[commandType] = handler
}

// registerDefaultHandlers registers built-in command handlers.
func (s *AgentService) registerDefaultHandlers() {
	s.RegisterCommandHandler("ping", func(ctx context.Context, cmd *db.Command) (map[string]interface{}, error) {
		return map[string]interface{}{
			"status": "pong",
			"time":   time.Now().Unix(),
		}, nil
	})
}
