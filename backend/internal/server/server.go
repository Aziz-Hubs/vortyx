package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"connectrpc.com/connect"

	v1 "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/v1"
	"github.com/abdul/vortyx/backend/gen/proto/go/vortyx/v1/vortyxv1connect"
	
	"github.com/abdul/vortyx/backend/internal/auth"

	// MSP
	controlPkg "github.com/abdul/vortyx/backend/internal/control"
	controlConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/control/v1/controlv1connect"
	gridPkg "github.com/abdul/vortyx/backend/internal/grid"
	gridConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/grid/v1/gridv1connect"
	horizonPkg "github.com/abdul/vortyx/backend/internal/horizon"
	horizonConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/horizon/v1/horizonv1connect"
	nexusPkg "github.com/abdul/vortyx/backend/internal/nexus"
	nexusConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/nexus/v1/nexusv1connect"
	opticPkg "github.com/abdul/vortyx/backend/internal/optic"
	opticConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/optic/v1/opticv1connect"
	pilotPkg "github.com/abdul/vortyx/backend/internal/pilot"
	pilotConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/pilot/v1/pilotv1connect"
	pulsePkg "github.com/abdul/vortyx/backend/internal/pulse"
	pulseConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/pulse/v1/pulsev1connect"

	// MSSP
	guardPkg "github.com/abdul/vortyx/backend/internal/guard"
	guardConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/guard/v1/guardv1connect"
	mindPkg "github.com/abdul/vortyx/backend/internal/mind"
	mindConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/mind/v1/mindv1connect"
	probePkg "github.com/abdul/vortyx/backend/internal/probe"
	probeConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/probe/v1/probev1connect"
	radarPkg "github.com/abdul/vortyx/backend/internal/radar"
	radarConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/radar/v1/radarv1connect"
	reflexPkg "github.com/abdul/vortyx/backend/internal/reflex"
	reflexConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/reflex/v1/reflexv1connect"
	shieldPkg "github.com/abdul/vortyx/backend/internal/shield"
	shieldConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/shield/v1/shieldv1connect"
	signalPkg "github.com/abdul/vortyx/backend/internal/signal"
	signalConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/signal/v1/signalv1connect"
	sonarPkg "github.com/abdul/vortyx/backend/internal/sonar"
	sonarConnect "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/sonar/v1/sonarv1connect"
)

type VortyxServer struct{}

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

func NewRouter(dbPool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Connect-Protocol-Version"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	
	// Auth Middleware
	r.Use(auth.AuthMiddleware)

	// Interceptors
	interceptors := connect.WithInterceptors(auth.NewAuthInterceptor())

	// Register Base Service
	path, handler := vortyxv1connect.NewVortyxServiceHandler(&VortyxServer{}, interceptors)
	r.Mount(path, handler)

	// MSP Services
	r.Mount(pulseConnect.NewPulseServiceHandler(pulsePkg.NewService(dbPool), interceptors))
	r.Mount(pilotConnect.NewPilotServiceHandler(pilotPkg.NewService(dbPool), interceptors))
	r.Mount(nexusConnect.NewNexusServiceHandler(nexusPkg.NewService(dbPool), interceptors))
	r.Mount(horizonConnect.NewHorizonServiceHandler(horizonPkg.NewService(dbPool), interceptors))
	r.Mount(controlConnect.NewControlServiceHandler(controlPkg.NewService(dbPool), interceptors))
	r.Mount(opticConnect.NewOpticServiceHandler(opticPkg.NewService(dbPool), interceptors))
	r.Mount(gridConnect.NewGridServiceHandler(gridPkg.NewService(dbPool), interceptors))

	// MSSP Services
	r.Mount(radarConnect.NewRadarServiceHandler(radarPkg.NewService(dbPool), interceptors))
	r.Mount(guardConnect.NewGuardServiceHandler(guardPkg.NewService(dbPool), interceptors))
	r.Mount(shieldConnect.NewShieldServiceHandler(shieldPkg.NewService(dbPool), interceptors))
	r.Mount(mindConnect.NewMindServiceHandler(mindPkg.NewService(dbPool), interceptors))
	r.Mount(probeConnect.NewProbeServiceHandler(probePkg.NewService(dbPool), interceptors))
	r.Mount(reflexConnect.NewReflexServiceHandler(reflexPkg.NewService(dbPool), interceptors))
	r.Mount(sonarConnect.NewSonarServiceHandler(sonarPkg.NewService(dbPool), interceptors))
	r.Mount(signalConnect.NewSignalServiceHandler(signalPkg.NewService(dbPool), interceptors))

	return h2c.NewHandler(r, &http2.Server{})
}
