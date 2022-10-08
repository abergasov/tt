package orchestrator

import (
	"context"
	"interview-fm-backend/internal/entities"
	"interview-fm-backend/internal/logger"
	"sync"

	"go.uber.org/zap"
)

// processSync receive request and process it synchronously. It will return only after all images are processed.
func (s *Service) processSync(ctx context.Context, request *entities.ResizeRequest) ([]entities.ResizeResult, error) {
	log := s.log.With(zap.Any("request", request)).
		With(zap.String("source", "request")).
		With(zap.Uint("width", request.Width)).
		With(zap.Uint("height", request.Height))

	log.Info("processing synchronous resizes")

	var wg sync.WaitGroup
	wg.Add(len(request.URLs))
	res := make(chan entities.ResizeResult, len(request.URLs))
	for _, url := range request.URLs {
		<-s.maxSyncImagesRequests // this will protect from too many parallel requests
		go func(imageURL string) {
			res <- s.processURL(ctx, log.With(zap.String("url", imageURL)), imageURL, request.Width, request.Height)
			wg.Done()
			s.maxSyncImagesRequests <- struct{}{} // release slot
		}(url)
	}
	wg.Wait()
	close(res)
	results := make([]entities.ResizeResult, 0, len(request.URLs))
	for result := range res {
		results = append(results, result)
	}
	return results, nil
}

// processURL process single image in separate goroutine and put result to channel
// if image already in cache - it just return it.
// else - make request to download data, resize it and put to cache
func (s *Service) processURL(ctx context.Context, log logger.AppLogger, url string, width, height uint) entities.ResizeResult {
	imageID := s.generateKey(url, width, height)
	newURL := s.imageURL(imageID)

	if s.cache.Contains(imageID) {
		log.Info("image already in cache")
		return entities.ResizeResult{
			URL:    newURL,
			Result: entities.ResizeResultStatusSuccess,
			Cached: true,
		}
	}

	log.Info("image not in cache, fetching and resizing")
	data, err := s.fetchAndResize(ctx, url, width, height)
	if err != nil {
		log.Error("failed to fetch and resize image", err)
		return entities.ResizeResult{Result: entities.ResizeResultStatusFailure}
	}
	s.cache.Add(imageID, data)
	return entities.ResizeResult{
		URL:    newURL,
		Result: entities.ResizeResultStatusSuccess,
		Cached: false,
	}
}
