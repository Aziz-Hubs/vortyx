package radar

import (
	"context"
	
	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	
	radarv1 "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/radar/v1"
	"github.com/abdul/vortyx/backend/internal/mssp/radar/db"
)

type Service struct {
	repo db.Querier
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		pool: pool,
		repo: db.New(pool),
	}
}

func (s *Service) GetStatus(
	ctx context.Context,
	req *connect.Request[radarv1.GetStatusRequest],
) (*connect.Response[radarv1.GetStatusResponse], error) {
	return connect.NewResponse(&radarv1.GetStatusResponse{
		Status:  "Operational",
		Version: "v1.0.0",
	}), nil
}
