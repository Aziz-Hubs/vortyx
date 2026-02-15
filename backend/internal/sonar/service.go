package sonar

import (
	"context"
	
	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	
	sonarv1 "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/sonar/v1"
	"github.com/abdul/vortyx/backend/internal/sonar/db"
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
	req *connect.Request[sonarv1.GetStatusRequest],
) (*connect.Response[sonarv1.GetStatusResponse], error) {
	return connect.NewResponse(&sonarv1.GetStatusResponse{
		Status:  "Operational",
		Version: "v1.0.0",
	}), nil
}
