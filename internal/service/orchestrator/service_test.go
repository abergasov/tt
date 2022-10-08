package orchestrator_test

import (
	"context"
	"fmt"
	"interview-fm-backend/internal/entities"
	"interview-fm-backend/internal/logger"
	"interview-fm-backend/internal/service/fetch"
	"interview-fm-backend/internal/service/orchestrator"
	"interview-fm-backend/internal/storage/cache"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type testFetcher struct {
	injectedFunc func() ([]byte, error)
}

func (t testFetcher) Fetch(_ context.Context, _ string) ([]byte, error) {
	return t.injectedFunc()
}

type testResizer struct {
}

func (t testResizer) ResizeImage(data []byte, _, _ uint) ([]byte, error) {
	return data, nil
}

const (
	baseURL       = "http://localhost:8080"
	sampleURL     = "http://localhost:8080/1/abc"
	sampleURLHash = "d2d069590f8f1b29101b851817358c6865070a078d362d8bd462db787e9f4d87"
)

var log, _ = logger.NewAppLogger()

func TestService_GetImage(t *testing.T) {
	t.Run("should return image", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		fetcher := fetch.NewMockFetcher(ctrl)
		cacheMock := cache.NewMockCacher(ctrl)

		service := orchestrator.NewService(baseURL, testResizer{}, fetcher, cacheMock, log)
		cacheMock.EXPECT().Contains("123").Return(true)
		cacheMock.EXPECT().Get(gomock.Any()).Return([]byte("123456"), true)
		res, ok, err := service.GetImage(context.Background(), "123")
		require.True(t, ok)
		require.NoError(t, err)
		require.Equal(t, []byte("123456"), res)
	})
	t.Run("should return error if image is in processing queue", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		cacheMock := cache.NewMockCacher(ctrl)
		fetcher := testFetcher{func() ([]byte, error) {
			cacheMock.EXPECT().Add(gomock.Any(), gomock.Any()) // check that all started requests are saved to cache before server stopped
			return []byte("123456"), nil
		}}

		service := orchestrator.NewService(baseURL, testResizer{}, fetcher, cacheMock, log)
		_, err := service.ProcessResizes(context.Background(), &entities.ResizeRequest{
			URLs:   []string{sampleURL},
			Height: 1,
			Width:  1,
		}, true)
		require.NoError(t, err)

		call := cacheMock.EXPECT().Contains(sampleURLHash).Return(false).Times(2)
		afterProcessing := cacheMock.EXPECT().Contains(sampleURLHash).Return(true).After(call)
		cacheMock.EXPECT().Get(sampleURLHash).Return([]byte("123456"), true).After(afterProcessing)

		res, ok, err := service.GetImage(context.Background(), sampleURLHash)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, []byte("123456"), res)
	})
}

func TestService_Shutdown(t *testing.T) {
	const imageProcess = 15
	testTimeout := time.After(5 * time.Second)
	ctrl := gomock.NewController(t)
	cacheMock := cache.NewMockCacher(ctrl)

	requestCounter := uint64(0)
	signalChan := make(chan struct{})
	fetcher := testFetcher{func() ([]byte, error) {
		atomic.AddUint64(&requestCounter, 1)
		cacheMock.EXPECT().Add(gomock.Any(), gomock.Any()) // check that all started requests are saved to cache before server stopped
		<-signalChan
		return []byte("123456"), nil
	}}
	service := orchestrator.NewService(baseURL, testResizer{}, fetcher, cacheMock, log)

	request := &entities.ResizeRequest{
		URLs:   make([]string, 0, imageProcess),
		Height: 1,
		Width:  1,
	}
	for i := 0; i < imageProcess; i++ {
		request.URLs = append(request.URLs, fmt.Sprintf("http://localhost:8080/%d/abc", i))
	}
	cacheMock.EXPECT().Contains(gomock.Any()).Return(false).AnyTimes()
	resp, err := service.ProcessResizes(context.Background(), request, true)
	require.NoError(t, err)
	require.True(t, len(resp) == imageProcess)

	doneChan := make(chan struct{})

	go func() {
		require.Eventually(t, func() bool {
			if atomic.LoadUint64(&requestCounter) != orchestrator.MaxAsyncAllowedRequests {
				return false
			}
			// started maximum of possible background tasks
			close(signalChan)
			return true
		}, 4*time.Second, 100*time.Millisecond)
		require.NoError(t, service.Shutdown())
		close(doneChan)
	}()

	select {
	case <-testTimeout:
		t.Fatal("Test didn't finish in time")
	case <-doneChan:
		break
	}
}
