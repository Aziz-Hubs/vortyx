package pulse

import (
	"context"
	
	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	
	pulsev1 "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/pulse/v1"
	"github.com/abdul/vortyx/backend/internal/msp/pulse/db"
)

type PulseService struct {
	repo db.Querier
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *PulseService {
	return &PulseService{
		pool: pool,
		repo: db.New(pool),
	}
}

func (s *PulseService) GetStatus(
	ctx context.Context,
	req *connect.Request[pulsev1.GetStatusRequest],
) (*connect.Response[pulsev1.GetStatusResponse], error) {
	return connect.NewResponse(&pulsev1.GetStatusResponse{
		Status:  "Operational",
		Version: "v1.0.0",
	}), nil
}
