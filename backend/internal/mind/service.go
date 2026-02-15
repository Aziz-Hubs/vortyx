package mind

import (
	"context"
	
	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	
	mindv1 "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/mind/v1"
	"github.com/abdul/vortyx/backend/internal/mind/db"
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
	req *connect.Request[mindv1.GetStatusRequest],
) (*connect.Response[mindv1.GetStatusResponse], error) {
	return connect.NewResponse(&mindv1.GetStatusResponse{
		Status:  "Operational",
		Version: "v1.0.0",
	}), nil
}
