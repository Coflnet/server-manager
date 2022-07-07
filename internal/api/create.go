package api

import (
	"net/http"
	"server-manager/internal/model"
	"server-manager/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func createServer(c *gin.Context) {

	request := &model.ServerRequest{}
	if err := c.ShouldBindJSON(request); err != nil {

		log.Error().Err(err).Msgf("could not parse body")
		c.AbortWithStatus(http.StatusUnprocessableEntity)

		return
	}

	s, err := usecase.RequestServerCreation(request)
	if err != nil {

		if re, ok := err.(*model.SlugDoesNotBelongToAServerType); ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, re.Error())
			return
		}

		if re, ok := err.(*model.ServerInvalidError); ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, re.Error())
			return
		}

		log.Error().Err(err).Msgf("could not create server")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusCreated, s)
}
