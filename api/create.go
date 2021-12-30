package api

import (
	"server-manager/iac"
	"server-manager/server"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func Create(c *fiber.Ctx) error {

	c.Accepts("application/json")
	serverType := new(server.ServerType)
	if err := c.BodyParser(serverType); err != nil {
		return err
	}

	if serverType.PlannedShtudown.Before(time.Now()) {
		log.Warn().Msgf("planned shutdown for server %s is before now %v", serverType.Type, serverType.PlannedShtudown)
	}

	log.Info().Msgf("creating a server of type %s", serverType.Type)
	serverType.CreatedAt = time.Now()

	serverType, err := iac.Create(serverType)

	if err != nil {
		return err
	}

	c.JSON(serverType)

	return nil
}
