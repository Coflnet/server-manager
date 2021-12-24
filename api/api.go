package api

import (
	"github.com/gofiber/fiber/v2"
)

func Start(errorCh chan<- error) {

	app := fiber.New()
	app.Post("/create", Create)
	app.Get("/list", List)
	app.Delete("/destroy", Destroy)
	app.Patch("/update", UpdatePlannedShutdown)

	errorCh <- app.Listen(":3000")
}
