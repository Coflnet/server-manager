package api

import (
	"net/http"
	"server-manager/internal/model"
	"server-manager/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func destroyServer(c *gin.Context) {

	var request *model.ServerRequest

	if err := c.BindJSON(&request); err != nil {
		log.Error().Err(err).Msgf("could not parse body destroy endpoint")
		c.AbortWithStatus(http.StatusUnprocessableEntity)
	}

	err := usecase.DestroyRequest(request)
	if err != nil {
		log.Error().Err(err).Msgf("could not destroy server")
		c.AbortWithStatus(http.StatusBadRequest)
	}

	c.Status(http.StatusOK)
}
