package grid

import (
	"context"
	
	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	
	gridv1 "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/grid/v1"
	"github.com/abdul/vortyx/backend/internal/msp/grid/db"
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
	req *connect.Request[gridv1.GetStatusRequest],
) (*connect.Response[gridv1.GetStatusResponse], error) {
	return connect.NewResponse(&gridv1.GetStatusResponse{
		Status:  "Operational",
		Version: "v1.0.0",
	}), nil
}
