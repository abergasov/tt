package main

import (
	"flag"
	"fmt"
	"interview-fm-backend/internal/logger"
	"interview-fm-backend/internal/routes"
	"interview-fm-backend/internal/service/fetch"
	"interview-fm-backend/internal/service/orchestrator"
	"interview-fm-backend/internal/service/resize"
	appCache "interview-fm-backend/internal/storage/cache"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

var appPort = flag.String("port", "8080", "App listen port")
var imageStorageHost = flag.String("imagehost", "http://localhost:8080", "Url to image storage service")

type Shutdowner interface {
	Shutdown() error
}

type ShutdownItem struct {
	Name    string
	Service Shutdowner
}

func main() {
	flag.Parse()
	log, err := logger.NewAppLogger()
	if err != nil {
		println(fmt.Errorf("error init logger: %w", err))
		return
	}

	cache, err := appCache.NewCache()
	if err != nil {
		log.Fatal("Failed to create cache", err)
	}

	resizer := orchestrator.NewService(*imageStorageHost, resize.NewResizerService(), fetch.NewService(), cache, log)
	app := routes.InitAppRouter(*appPort, resizer)
	go func() {
		log.Info("starting service", zap.String("port", *appPort))
		if err = app.Run(); err != nil {
			log.Fatal("error start service", err)
		}
	}()

	// register app shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c // This blocks the main thread until an interrupt is received

	// not using context, because order is important
	// 1. stop router, stop accept new requests
	// 2. stop resizer and save cache
	// 3. stop cache and dump data
	// 4. exit app
	shutdownItems := []ShutdownItem{
		{"router", app},
		{"resizer", resizer},
		{"cache", cache},
	}

	for i := range shutdownItems {
		log.Info("shutdown service", zap.String("service", shutdownItems[i].Name))
		if err = shutdownItems[i].Service.Shutdown(); err != nil {
			log.Fatal("error shutdown service", err, zap.String("service", shutdownItems[i].Name))
		}
	}
	log.Info("app was successful shutdown")
}
