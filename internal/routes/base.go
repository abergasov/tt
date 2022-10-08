package routes

import (
	"interview-fm-backend/internal/service/orchestrator"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type AppRouter struct {
	service  orchestrator.Orchestrator
	appPort  string
	fiberApp *fiber.App
}

// InitAppRouter initializes the app router.
func InitAppRouter(appPort string, service orchestrator.Orchestrator) *AppRouter {
	fiberApp := fiber.New(
		fiber.Config{
			DisableStartupMessage: true,
			BodyLimit:             8 * 1024,
		},
	)

	fiberApp.Use(recover.New())

	app := &AppRouter{
		appPort:  appPort,
		fiberApp: fiberApp,
		service:  service,
	}
	app.initRoutes()
	return app
}

func (a *AppRouter) initRoutes() {
	a.fiberApp.Get("/ping", func(ctx *fiber.Ctx) error {
		return ctx.SendString("pong")
	})
	a.fiberApp.Post("/v1/resize", a.resize)
	a.fiberApp.Get("/v1/image/:image.jpg", a.getImage)
}

// Run starts the server.
func (a *AppRouter) Run() error {
	return a.fiberApp.Listen(":" + a.appPort)
}

// Shutdown gracefully shuts down the server.
func (a *AppRouter) Shutdown() error {
	return a.fiberApp.Shutdown()
}
