package pilot

import (
	"context"
	
	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	
	pilotv1 "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/pilot/v1"
	"github.com/abdul/vortyx/backend/internal/msp/pilot/db"
)

type PilotService struct {
	repo db.Querier
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *PilotService {
	return &PilotService{
		pool: pool,
		repo: db.New(pool),
	}
}

func (s *PilotService) GetStatus(
	ctx context.Context,
	req *connect.Request[pilotv1.GetStatusRequest],
) (*connect.Response[pilotv1.GetStatusResponse], error) {
	return connect.NewResponse(&pilotv1.GetStatusResponse{
		Status:  "Operational",
		Version: "v1.0.0",
	}), nil
}
