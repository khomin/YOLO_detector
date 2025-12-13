package route

import (
	"tuna-backend/bootstrap"

	"example.com/app/controller"
	"github.com/gin-gonic/gin"
)

func Setup(env *bootstrap.Env,
	db *bootstrap.DatabaseUseCase,
	gMain *gin.Engine,
) {
	lc := &controller.ClientControlller{
		Env: env,
	}

	route := gMain.Group("")

	route.POST("/v1/track_event", lc.OnTrackUpdate)
	route.POST("/v1/test", lc.TestMethod)
}
