// Package server provides the main HTTP server implementation for the Vortyx backend.
// It sets up the chi router, configures interceptors, and registers all service handlers.
//
// The server uses ConnectRPC for gRPC-style APIs and chi for HTTP routing.
// Authentication is handled by the auth package using Zitadel OIDC.
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"connectrpc.com/connect"

	v1 "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/v1"
	"github.com/abdul/vortyx/backend/gen/proto/go/vortyx/v1/vortyxv1connect"
	platformConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/platform/v1/platformv1connect"

	"github.com/abdul/vortyx/backend/internal/server/health"
	"github.com/abdul/vortyx/backend/internal/server/interceptors"
	platformPkg "github.com/abdul/vortyx/backend/internal/platform"

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

	// VORT Agent package
	vortConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/vort/v1/vortv1connect"
	vortPkg "github.com/abdul/vortyx/backend/internal/vort/service"
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
	// Base Interceptors
	// ==========================
	r.Use(interceptors.CORSMiddleware(interceptors.DefaultCORSConfig()))
	r.Use(interceptors.RequestIDMiddleware())
	r.Use(interceptors.TracingMiddleware("vortyx-backend"))
	r.Use(interceptors.LoggerMiddleware())
	r.Use(interceptors.RecovererMiddleware())
	r.Use(interceptors.SecurityMiddleware(interceptors.DefaultSecurityConfig()))
	r.Use(interceptors.CompressionMiddleware())
	r.Use(interceptors.MetricsMiddleware())
	r.Use(interceptors.RateLimitMiddleware(interceptors.DefaultRateLimitConfig()))

	// ==========================
	// Security Interceptors
	// ==========================
	authInterceptor, connectInterceptors := interceptors.AuthInterceptor(interceptors.DefaultAuthConfig())
	if authInterceptor != nil {
		r.Use(authInterceptor)
		r.Use(interceptors.AuditMiddleware())
	}

	// Register the base Vortyx service handler.
	path, handler := vortyxv1connect.NewVortyxServiceHandler(&VortyxServer{}, connectInterceptors)
	r.Mount(path, handler)

	// Health Check
	r.Get("/health", health.NewHandler(dbPool))

	projectID := os.Getenv("ZITADEL_API_PROJECT_ID")
	if projectID == "" {
		projectID = os.Getenv("ZITADEL_PLATFORM_PROJECT_ID")
	}

	zitadelMgmtClient, err := NewZitadelManagementClient(context.Background())
	if err != nil {
		zitadelMgmtClient = nil
	}

	platformService := platformPkg.NewService(dbPool, zitadelMgmtClient, projectID)
	r.Mount(platformConnect.NewPlatformServiceHandler(platformService, connectInterceptors))

	// Register MSP (Managed Service Provider) services.
	r.Mount(mspPulseConnect.NewPulseServiceHandler(mspPulsePkg.NewService(dbPool), connectInterceptors))
	r.Mount(mspPilotConnect.NewPilotServiceHandler(mspPilotPkg.NewService(dbPool), connectInterceptors))
	r.Mount(mspNexusConnect.NewNexusServiceHandler(mspNexusPkg.NewService(dbPool), connectInterceptors))
	r.Mount(mspHorizonConnect.NewHorizonServiceHandler(mspHorizonPkg.NewService(dbPool), connectInterceptors))
	r.Mount(mspControlConnect.NewControlServiceHandler(mspControlPkg.NewService(dbPool), connectInterceptors))
	r.Mount(mspOpticConnect.NewOpticServiceHandler(mspOpticPkg.NewService(dbPool), connectInterceptors))
	r.Mount(mspGridConnect.NewGridServiceHandler(mspGridPkg.NewService(dbPool), connectInterceptors))

	// Register MSSP (Managed Security Service Provider) services.
	r.Mount(msspRadarConnect.NewRadarServiceHandler(msspRadarPkg.NewService(dbPool), connectInterceptors))
	r.Mount(msspGuardConnect.NewGuardServiceHandler(msspGuardPkg.NewService(dbPool), connectInterceptors))
	r.Mount(msspShieldConnect.NewShieldServiceHandler(msspShieldPkg.NewService(dbPool), connectInterceptors))
	r.Mount(msspMindConnect.NewMindServiceHandler(msspMindPkg.NewService(dbPool), connectInterceptors))
	r.Mount(msspProbeConnect.NewProbeServiceHandler(msspProbePkg.NewService(dbPool), connectInterceptors))
	r.Mount(msspReflexConnect.NewReflexServiceHandler(msspReflexPkg.NewService(dbPool), connectInterceptors))
	r.Mount(msspSonarConnect.NewSonarServiceHandler(msspSonarPkg.NewService(dbPool), connectInterceptors))
	r.Mount(msspSignalConnect.NewSignalServiceHandler(msspSignalPkg.NewService(dbPool), connectInterceptors))

	// Register VORT Agent service.
	// Note: We use the Zitadel-based authentication interceptor.
	vortService := vortPkg.NewServiceFromPool(dbPool)
	r.Mount(vortConnect.NewVortServiceHandler(vortService, connectInterceptors))

	// Register profiling handlers (for debugging)
	// Only enable in development mode
	if os.Getenv("ENV") != "production" && os.Getenv("ENV") != "prod" {
		interceptors.MountPprof(r)
	}

	// Wrap with h2c to support HTTP/2 without TLS.
	// This enables gRPC-style streaming and better performance.
	return h2c.NewHandler(r, &http2.Server{})
}
