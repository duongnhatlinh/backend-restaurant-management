package routes

import (
	"github.com/gin-gonic/gin"
	controller "restaurant_management/controllers"
)

func UserRoutes(routes *gin.Engine) {
	routes.GET("/users", controller.GetUsers())
	routes.GET("/users/:id", controller.GetUser())
	routes.POST("/users/signup", controller.SignUp())
	routes.POST("/users/login", controller.LogIn())
}
