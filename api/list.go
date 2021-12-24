package api

import (
	"server-manager/mongo"
	"server-manager/server"

	"github.com/gofiber/fiber/v2"
)

func List(c *fiber.Ctx) error {

	servers, err := mongo.List()
	if err != nil {
		return err
	}

	// change from nil to empty array
	if servers == nil {
		servers = []*server.ServerType{}
	}
	c.JSON(servers)

	return nil
}
