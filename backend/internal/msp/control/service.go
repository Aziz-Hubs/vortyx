package control

import (
	"context"
	
	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	
	controlv1 "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/control/v1"
	"github.com/abdul/vortyx/backend/internal/msp/control/db"
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
	req *connect.Request[controlv1.GetStatusRequest],
) (*connect.Response[controlv1.GetStatusResponse], error) {
	return connect.NewResponse(&controlv1.GetStatusResponse{
		Status:  "Operational",
		Version: "v1.0.0",
	}), nil
}
