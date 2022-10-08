package orchestrator

import (
	"context"
	"interview-fm-backend/internal/entities"

	"go.uber.org/zap"
)

// processAsync receive request and put it to queue. It will return immediately with status "processing".
func (s *Service) processAsync(_ context.Context, request *entities.ResizeRequest) ([]entities.ResizeResult, error) {
	log := s.log.With(zap.Uint("width", request.Width)).
		With(zap.Uint("height", request.Height))

	results := make([]entities.ResizeResult, 0, len(request.URLs))
	for _, url := range request.URLs {
		imageID := s.generateKey(url, request.Width, request.Height)
		newURL := s.imageURL(imageID)
		results = append(results, entities.ResizeResult{
			URL:    newURL,
			Result: entities.ResizeResultStatusProcessing,
			Cached: true,
		})
		jobLog := log.With(zap.String("url", url)).With(zap.String("imageID", imageID))
		s.handleNewJob(jobLog, url, imageID, request.Width, request.Height)
	}
	return results, nil
}
