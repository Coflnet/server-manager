package api

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Start() error {

	r := gin.Default()

	r.Use(JSONMiddleware)

	r.GET("/list", listActiveServers)
	r.POST("/create", createServer)
	r.DELETE("/destroy", destroyServer)

	r.POST("/extend", extendServer)

	r.GET("/status", status)

	log.Info().Msg("starting api")
	return r.Run()
}

func JSONMiddleware(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Next()
}
