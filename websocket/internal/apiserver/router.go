package apiserver

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"socket-demo/internal/apiserver/controller/home"
	"socket-demo/internal/apiserver/controller/user"
)

func initRouter(g *gin.Engine) {
	installMiddleware(g)
	installController(g)
}

func installMiddleware(g *gin.Engine) {
	return
}

func installController(g *gin.Engine) *gin.Engine {

	g.GET("/healthz", func(context *gin.Context) {
		context.JSON(http.StatusOK, nil)
	})

	g.POST("/login", func(context *gin.Context) {
		context.JSON(http.StatusOK, map[string]string{"msg": "success"})
	})

	v1 := g.Group("/v1")

	{

		homev1 := v1.Group("/home")
		{
			homeController := home.NewHomeController()
			homev1.GET("", homeController.Data)
		}

		userv1 := v1.Group("/users")
		{
			userController := user.NewUserController()
			userv1.GET("", userController.Get)
		}

	}

	return g
}
