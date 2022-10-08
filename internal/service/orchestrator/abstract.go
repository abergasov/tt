package orchestrator

import (
	"context"
	"interview-fm-backend/internal/entities"
)

type Orchestrator interface {
	ProcessResizes(ctx context.Context, request *entities.ResizeRequest, async bool) ([]entities.ResizeResult, error)
	GetImage(ctx context.Context, imageID string) ([]byte, bool, error)
	Shutdown() error
}
