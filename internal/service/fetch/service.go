package fetch

import (
	"context"
	"interview-fm-backend/internal/utils"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Fetch(ctx context.Context, url string) ([]byte, error) {
	return utils.FetchURL(ctx, url)
}
