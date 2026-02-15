package probe

import (
	"context"
	
	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	
	probev1 "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/probe/v1"
	"github.com/abdul/vortyx/backend/internal/probe/db"
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
	req *connect.Request[probev1.GetStatusRequest],
) (*connect.Response[probev1.GetStatusResponse], error) {
	return connect.NewResponse(&probev1.GetStatusResponse{
		Status:  "Operational",
		Version: "v1.0.0",
	}), nil
}
