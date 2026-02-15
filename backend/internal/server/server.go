// Package server provides the main HTTP server implementation for the Vortyx backend.
// It sets up the chi router, configures middleware, and registers all service handlers.
//
// The server uses ConnectRPC for gRPC-style APIs and chi for HTTP routing.
// Authentication is handled by the auth package using Zitadel OIDC.
package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"connectrpc.com/connect"

	v1 "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/v1"
	"github.com/abdul/vortyx/backend/gen/proto/go/vortyx/v1/vortyxv1connect"

	middleware "github.com/abdul/vortyx/backend/internal/server/middleware"

	// MSP (Managed Service Provider) packages
	mspControlPkg "github.com/abdul/vortyx/backend/internal/msp/control"
	mspGridPkg "github.com/abdul/vortyx/backend/internal/msp/grid"
	mspHorizonPkg "github.com/abdul/vortyx/backend/internal/msp/horizon"
	mspNexusPkg "github.com/abdul/vortyx/backend/internal/msp/nexus"
	mspOpticPkg "github.com/abdul/vortyx/backend/internal/msp/optic"
	mspPilotPkg "github.com/abdul/vortyx/backend/internal/msp/pilot"
	mspPulsePkg "github.com/abdul/vortyx/backend/internal/msp/pulse"

	mspControlConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/control/v1/controlv1connect"
	mspGridConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/grid/v1/gridv1connect"
	mspHorizonConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/horizon/v1/horizonv1connect"
	mspNexusConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/nexus/v1/nexusv1connect"
	mspOpticConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/optic/v1/opticv1connect"
	mspPilotConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/pilot/v1/pilotv1connect"
	mspPulseConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/pulse/v1/pulsev1connect"

	// MSSP (Managed Security Service Provider) packages
	msspGuardPkg "github.com/abdul/vortyx/backend/internal/mssp/guard"
	msspMindPkg "github.com/abdul/vortyx/backend/internal/mssp/mind"
	msspProbePkg "github.com/abdul/vortyx/backend/internal/mssp/probe"
	msspRadarPkg "github.com/abdul/vortyx/backend/internal/mssp/radar"
	msspReflexPkg "github.com/abdul/vortyx/backend/internal/mssp/reflex"
	msspShieldPkg "github.com/abdul/vortyx/backend/internal/mssp/shield"
	msspSignalPkg "github.com/abdul/vortyx/backend/internal/mssp/signal"
	msspSonarPkg "github.com/abdul/vortyx/backend/internal/mssp/sonar"

	msspGuardConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/guard/v1/guardv1connect"
	msspMindConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/mind/v1/mindv1connect"
	msspProbeConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/probe/v1/probev1connect"
	msspRadarConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/radar/v1/radarv1connect"
	msspReflexConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/reflex/v1/reflexv1connect"
	msspShieldConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/shield/v1/shieldv1connect"
	msspSignalConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/signal/v1/signalv1connect"
	msspSonarConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/sonar/v1/sonarv1connect"

	// Platform package
	// Note: Platform service is planned for future implementation
)

// VortyxServer is the main server implementation that handles base API requests.
// It implements the VortyxServiceHandler interface defined in the protobuf.
type VortyxServer struct{}

// Ping handles the base ping request for health checking.
// This endpoint does not require authentication and can be used
// to verify the server is responding.
func (s *VortyxServer) Ping(
	ctx context.Context,
	req *connect.Request[v1.PingRequest],
) (*connect.Response[v1.PingResponse], error) {
	res := connect.NewResponse(&v1.PingResponse{
		Message: fmt.Sprintf("Hello, %s!", req.Msg.Message),
	})
	res.Header().Set("Vortyx-Version", "v1")
	return res, nil
}

// NewRouter creates and configures the main HTTP router with all middleware
// and service handlers. This is the entry point for setting up the server's
// routing layer.
//
// The router is configured with:
//   - Zitadel authentication middleware
//   - Chi logging and recovery middleware
//   - CORS handling for frontend access
//   - All MSP and MSSP service handlers
//
// The returned handler uses h2c (HTTP/2 over cleartext) to support
// both HTTP/1.1 and HTTP/2 clients without TLS termination.
func NewRouter(dbPool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	// ==========================
	// Security Middleware
	// ==========================
	authMiddleware, interceptors := middleware.AuthMiddleware(middleware.DefaultAuthConfig())
	if authMiddleware != nil {
		r.Use(authMiddleware)
	}

	// ==========================
	// Core Middleware
	// ==========================
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RecovererMiddleware())

	// ==========================
	// HTTP Middleware
	// ==========================
	r.Use(middleware.CORSMiddleware(middleware.DefaultCORSConfig()))

	// Register the base Vortyx service handler.
	path, handler := vortyxv1connect.NewVortyxServiceHandler(&VortyxServer{}, interceptors)
	r.Mount(path, handler)

	// Register MSP (Managed Service Provider) services.
	r.Mount(mspPulseConnect.NewPulseServiceHandler(mspPulsePkg.NewService(dbPool), interceptors))
	r.Mount(mspPilotConnect.NewPilotServiceHandler(mspPilotPkg.NewService(dbPool), interceptors))
	r.Mount(mspNexusConnect.NewNexusServiceHandler(mspNexusPkg.NewService(dbPool), interceptors))
	r.Mount(mspHorizonConnect.NewHorizonServiceHandler(mspHorizonPkg.NewService(dbPool), interceptors))
	r.Mount(mspControlConnect.NewControlServiceHandler(mspControlPkg.NewService(dbPool), interceptors))
	r.Mount(mspOpticConnect.NewOpticServiceHandler(mspOpticPkg.NewService(dbPool), interceptors))
	r.Mount(mspGridConnect.NewGridServiceHandler(mspGridPkg.NewService(dbPool), interceptors))

	// Register MSSP (Managed Security Service Provider) services.
	r.Mount(msspRadarConnect.NewRadarServiceHandler(msspRadarPkg.NewService(dbPool), interceptors))
	r.Mount(msspGuardConnect.NewGuardServiceHandler(msspGuardPkg.NewService(dbPool), interceptors))
	r.Mount(msspShieldConnect.NewShieldServiceHandler(msspShieldPkg.NewService(dbPool), interceptors))
	r.Mount(msspMindConnect.NewMindServiceHandler(msspMindPkg.NewService(dbPool), interceptors))
	r.Mount(msspProbeConnect.NewProbeServiceHandler(msspProbePkg.NewService(dbPool), interceptors))
	r.Mount(msspReflexConnect.NewReflexServiceHandler(msspReflexPkg.NewService(dbPool), interceptors))
	r.Mount(msspSonarConnect.NewSonarServiceHandler(msspSonarPkg.NewService(dbPool), interceptors))
	r.Mount(msspSignalConnect.NewSignalServiceHandler(msspSignalPkg.NewService(dbPool), interceptors))

	// Wrap with h2c to support HTTP/2 without TLS.
	// This enables gRPC-style streaming and better performance.
	return h2c.NewHandler(r, &http2.Server{})
}
