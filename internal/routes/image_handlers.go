package routes

import (
	"interview-fm-backend/internal/entities"

	"github.com/gofiber/fiber/v2"
)

func (a *AppRouter) resize(ctx *fiber.Ctx) error {
	var resizeRequest *entities.ResizeRequest
	if err := ctx.BodyParser(&resizeRequest); err != nil {
		return fiber.ErrBadRequest
	}
	var asyncProcess bool
	if ctx.Query("async") == "true" {
		asyncProcess = true
	}
	response, err := a.service.ProcessResizes(ctx.UserContext(), resizeRequest, asyncProcess)
	if err != nil {
		return fiber.ErrInternalServerError
	}
	return ctx.JSON(response)
}

func (a *AppRouter) getImage(ctx *fiber.Ctx) error {
	data, ok, err := a.service.GetImage(ctx.UserContext(), ctx.Params("image"))
	if err != nil {
		return fiber.ErrInternalServerError
	}
	if !ok {
		return fiber.ErrNotFound
	}
	ctx.Set("Content-Type", "image/jpeg")
	return ctx.Send(data)
}
