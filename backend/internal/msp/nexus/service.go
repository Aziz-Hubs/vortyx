package nexus

import (
	"context"
	
	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	
	nexusv1 "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/nexus/v1"
	"github.com/abdul/vortyx/backend/internal/msp/nexus/db"
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
	req *connect.Request[nexusv1.GetStatusRequest],
) (*connect.Response[nexusv1.GetStatusResponse], error) {
	return connect.NewResponse(&nexusv1.GetStatusResponse{
		Status:  "Operational",
		Version: "v1.0.0",
	}), nil
}
