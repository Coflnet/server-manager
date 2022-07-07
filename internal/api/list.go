package api

import (
	"net/http"
	"server-manager/internal/model"
	"server-manager/internal/mongo"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func listActiveServers(c *gin.Context) {

	servers, err := mongo.ListActiveServers()
	if err != nil {
		log.Error().Err(err).Msgf("could not list servers")
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	// change from nil to empty array
	if servers == nil {
		servers = []*model.Server{}
	}

	c.JSON(http.StatusOK, servers)
}
