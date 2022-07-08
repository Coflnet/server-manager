package api

import (
	"github.com/gin-gonic/gin"
	"server-manager/internal/usecase"
)

func status(c *gin.Context) {

	activeServers, err := usecase.ActiveServers()
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	activeHetznerServers, err := usecase.ActiveHetznerServers()
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	activeGoogleServers, err := usecase.ActiveGoogleServers()
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"activeServers":        activeServers,
		"activeHetznerServers": activeHetznerServers,
		"activeGoogleServers":  activeGoogleServers,
	})
}
