package guard

import (
	"context"
	
	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	
	guardv1 "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/guard/v1"
	"github.com/abdul/vortyx/backend/internal/guard/db"
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
	req *connect.Request[guardv1.GetStatusRequest],
) (*connect.Response[guardv1.GetStatusResponse], error) {
	return connect.NewResponse(&guardv1.GetStatusResponse{
		Status:  "Operational",
		Version: "v1.0.0",
	}), nil
}
