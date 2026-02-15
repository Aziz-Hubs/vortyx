package signal

import (
	"context"
	
	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	
	signalv1 "github.com/abdul/vortyx/backend/gen/proto/go/vortyx/signal/v1"
	"github.com/abdul/vortyx/backend/internal/mssp/signal/db"
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
	req *connect.Request[signalv1.GetStatusRequest],
) (*connect.Response[signalv1.GetStatusResponse], error) {
	return connect.NewResponse(&signalv1.GetStatusResponse{
		Status:  "Operational",
		Version: "v1.0.0",
	}), nil
}
