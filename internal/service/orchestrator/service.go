package orchestrator

import (
	"container/list"
	"context"
	"fmt"
	"interview-fm-backend/internal/entities"
	"interview-fm-backend/internal/logger"
	"interview-fm-backend/internal/service/fetch"
	"interview-fm-backend/internal/service/resize"
	"interview-fm-backend/internal/storage/cache"
	"interview-fm-backend/internal/utils"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	MaxAllowedRequests      = 10
	MaxAsyncAllowedRequests = 10
)

type imageStatusContainer struct {
	status entities.ResizeResultStatus
	signal chan struct{}
}

type Service struct {
	baseURL                string
	cache                  cache.Cacher
	log                    logger.AppLogger
	resizer                resize.Resizer
	fetcherService         fetch.Fetcher
	maxSyncImagesRequests  chan struct{} // how much parallel execution allowed
	maxAsyncImagesRequests chan struct{} // how much parallel execution allowed for async processing

	queue   *list.List // list of tasks to process in async
	queueMU sync.Mutex

	imageStatus   map[string]*imageStatusContainer // map of imageID to trace status
	imageStatusMU sync.RWMutex

	ctx        context.Context // context for graceful shutdown
	cancel     context.CancelFunc
	workerDone chan struct{} // channel to notify that background worker is done and service stopped
}

func NewService(baseURL string, resizer resize.Resizer, fetcherService fetch.Fetcher, cache cache.Cacher, log logger.AppLogger) *Service {
	srv := &Service{
		resizer:                resizer,
		fetcherService:         fetcherService,
		cache:                  cache,
		baseURL:                baseURL,
		log:                    log.With(zap.String("service", "resize")),
		maxSyncImagesRequests:  make(chan struct{}, MaxAllowedRequests),
		maxAsyncImagesRequests: make(chan struct{}, MaxAsyncAllowedRequests),

		queue:   list.New(),
		queueMU: sync.Mutex{},

		imageStatus:   map[string]*imageStatusContainer{},
		imageStatusMU: sync.RWMutex{},
		workerDone:    make(chan struct{}),
	}
	for i := 0; i < MaxAllowedRequests; i++ {
		srv.maxSyncImagesRequests <- struct{}{}
	}
	for i := 0; i < MaxAsyncAllowedRequests; i++ {
		srv.maxAsyncImagesRequests <- struct{}{}
	}

	srv.ctx, srv.cancel = context.WithCancel(context.Background())
	go srv.worker()
	return srv
}

// ProcessResizes process resize requests in sync or async mode, depending from `async` flag.
func (s *Service) ProcessResizes(ctx context.Context, request *entities.ResizeRequest, async bool) ([]entities.ResizeResult, error) {
	if async {
		return s.processAsync(ctx, request)
	}
	return s.processSync(ctx, request)
}

func (s *Service) fetchAndResize(ctx context.Context, url string, width, height uint) ([]byte, error) {
	data, err := s.fetcherService.Fetch(ctx, url)
	if err != nil {
		return nil, err
	}
	return s.resizer.ResizeImage(data, width, height)
}

// GetImage in current realization just returns image from in-memory.
// for future enhancement added context, in case expected network request for external cache service.
// If image not found in cache - than also check in imageStatus map. If failed - serve 404 error.
// Then start wait for end of processing.
func (s *Service) GetImage(ctx context.Context, imageID string) ([]byte, bool, error) {
	log := s.log.With(zap.String("method", "GetImage")).With(zap.String("image_id", imageID))
	log.Info("getting image")
	if s.cache.Contains(imageID) {
		imgData, ok := s.cache.Get(imageID)
		return imgData, ok, nil
	}
	log.Info("image not found in cache")
	s.imageStatusMU.RLock()
	container, ok := s.imageStatus[imageID]
	s.imageStatusMU.RUnlock()
	if !ok {
		log.Info("image not found in processing queue")
		return nil, false, nil
	}
	if container.status == entities.ResizeResultStatusFailure {
		log.Info("image processing failed")
		return nil, false, nil
	}
	log.Info("image is processing, wait to finish")

	ctxT, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	for {
		select {
		case <-container.signal:
			log.Info("image processing finished")
			if s.cache.Contains(imageID) {
				log.Info("image found in cache after processing")
				imgData, ok := s.cache.Get(imageID)
				return imgData, ok, nil
			}
			log.Info("image not found in cache after processing")
			return nil, false, nil
		case <-ctxT.Done():
			return nil, false, ctxT.Err()
		}
	}
}

// generateKey calculate hash from string, width and height
// It is allows store in cache same image for different sizes
func (s *Service) generateKey(url string, width, height uint) string {
	return utils.GenerateKey(fmt.Sprintf("%s_%d_%d", url, width, height))
}

func (s *Service) imageURL(imageID string) string {
	return fmt.Sprintf("%s/v1/image/%s.jpg", s.baseURL, imageID)
}

// Shutdown gracefully shutdown service.
// First stop starting new tasks
// Than wait for current executing tasks are done
// Than store current queue in dump file and exit
func (s *Service) Shutdown() error {
	s.cancel()
	<-s.workerDone
	s.dumpQueue()
	return nil
}

func (s *Service) dumpQueue() {
	s.log.Info("dumping queue...")
	s.log.Info("dumping queue done")
}
