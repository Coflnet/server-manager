package api

import (
	"github.com/gin-gonic/gin"
	"server-manager/internal/mongo"
	"server-manager/internal/usecase"
	"time"
)

type ServerExtensionRequest struct {
	Name          string `json:"name"`
	ExtendSeconds int    `json:"extendSeconds"`
}

func extendServer(c *gin.Context) {

	var request ServerExtensionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	duration := time.Duration(request.ExtendSeconds) * time.Second

	server, err := mongo.ServerByName(request.Name)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = usecase.ExtendServer(server, duration)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "server extended",
	})
}
