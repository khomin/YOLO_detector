package controller

import (
	"net/http"

	"example.com/app/bootstrap"
	"github.com/gin-gonic/gin"
)

type ClientControlller struct {
	Env *bootstrap.Env
}

func (cc *ClientControlller) TestMethod(c *gin.Context) {
	response := map[string]interface{}{
		"success": true,
	}
	c.JSON(http.StatusOK, response)
}

func (cc *ClientControlller) OnTrackUpdate(c *gin.Context) {
	response := map[string]interface{}{
		"success": true,
	}
	c.JSON(http.StatusOK, response)
}
