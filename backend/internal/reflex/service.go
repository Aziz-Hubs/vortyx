package reflex

import (
	"context"
	
	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	
	reflexv1 "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/reflex/v1"
	"github.com/abdul/vortyx/backend/internal/reflex/db"
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
	req *connect.Request[reflexv1.GetStatusRequest],
) (*connect.Response[reflexv1.GetStatusResponse], error) {
	return connect.NewResponse(&reflexv1.GetStatusResponse{
		Status:  "Operational",
		Version: "v1.0.0",
	}), nil
}
