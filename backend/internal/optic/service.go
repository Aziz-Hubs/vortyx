package optic

import (
	"context"
	
	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	
	opticv1 "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/optic/v1"
	"github.com/abdul/vortyx/backend/internal/optic/db"
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
	req *connect.Request[opticv1.GetStatusRequest],
) (*connect.Response[opticv1.GetStatusResponse], error) {
	return connect.NewResponse(&opticv1.GetStatusResponse{
		Status:  "Operational",
		Version: "v1.0.0",
	}), nil
}
