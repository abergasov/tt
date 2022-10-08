package orchestrator

import (
	"container/list"
	"context"
	"fmt"
	"interview-fm-backend/internal/entities"
	"interview-fm-backend/internal/logger"
	"sync"
	"time"

	"go.uber.org/zap"
)

type task struct {
	url     string
	imageID string
	width   uint
	height  uint
}

// handleNewJob save value `imageURLHash` at map with status "processing" and add new task to queue.
// If same hash already in map - than it mean that image already in queue, so we just return.
// If processing return error - we update map with status "failed".
// If processing return success - we update map with status "success".
func (s *Service) handleNewJob(log logger.AppLogger, url, imageID string, width, height uint) {
	log.Info("handling new job")
	s.imageStatusMU.Lock()
	defer s.imageStatusMU.Unlock()
	if _, ok := s.imageStatus[imageID]; ok {
		log.Info("image already in progress")
		return
	}
	s.imageStatus[imageID] = &imageStatusContainer{
		status: entities.ResizeResultStatusProcessing,
		signal: make(chan struct{}),
	}

	s.queueMU.Lock()
	defer s.queueMU.Unlock()
	s.queue.PushBack(&task{
		url:     url,
		imageID: imageID,
		height:  height,
		width:   width,
	})
	log.Info("new job added to queue")
}

// worker start infinite loop to process queue. It will stop when context is done and close workerDone channel at the end.
// When from maxAsyncImagesRequests was read empty struct - than task processing starting from queue.
// When task is done, it will push struct back to maxAsyncImagesRequests.
// So maxAsyncImagesRequests regulate, how much parallel execution allowed for async processing.
// When loop is done, it will wait all tasks to be done, using sync.WaitGroup to control it.
func (s *Service) worker() {
	var wg sync.WaitGroup
	var c sync.Cond
	c.Broadcast()
loop:
	for {
		select {
		case <-s.maxAsyncImagesRequests:
			wg.Add(1)
			s.queueMU.Lock()
			e := s.queue.Front()
			if e != nil {
				s.queue.Remove(e)
			}
			s.queueMU.Unlock()
			if e == nil {
				s.maxAsyncImagesRequests <- struct{}{}
				wg.Done()
				// no more tasks in queue, wait for new tasks
				time.Sleep(50 * time.Millisecond)
				continue
			}
			go func(el *list.Element) {
				s.processQueue(el)
				s.maxAsyncImagesRequests <- struct{}{}
				wg.Done()
			}(e)
		case <-s.ctx.Done():
			break loop
		}
	}
	wg.Wait()
	close(s.workerDone)
}

func (s *Service) processQueue(el *list.Element) {
	t, ok := el.Value.(*task)
	if !ok {
		s.log.Error("invalid task type", fmt.Errorf("expected type is pointer to task"), zap.Any("task", el))
		return
	}

	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	log := s.log.With(zap.String("source", "background")).
		With(zap.Uint("width", t.width)).
		With(zap.Uint("height", t.height)).
		With(zap.String("url", t.url))

	log.Info("processing background resizes")
	res := s.processURL(ctx, log, t.url, t.width, t.height)
	log.Info("background resizes done")
	s.imageStatusMU.Lock()
	defer s.imageStatusMU.Unlock()
	close(s.imageStatus[t.imageID].signal)
	s.imageStatus[t.imageID].status = res.Result
}
