package shield

import (
	"context"
	
	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	
	shieldv1 "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/shield/v1"
	"github.com/abdul/vortyx/backend/internal/mssp/shield/db"
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
	req *connect.Request[shieldv1.GetStatusRequest],
) (*connect.Response[shieldv1.GetStatusResponse], error) {
	return connect.NewResponse(&shieldv1.GetStatusResponse{
		Status:  "Operational",
		Version: "v1.0.0",
	}), nil
}
