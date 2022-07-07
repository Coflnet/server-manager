package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func extendServer(c *gin.Context) {
	c.Status(http.StatusNotImplemented)
}
