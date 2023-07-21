package routes

import (
	"github.com/gin-gonic/gin"
	controller "restaurant_management/controllers"
)

func FoodRoutes(routes *gin.Engine) {
	routes.GET("/foods", controller.GetFoods())
	routes.GET("/foods/:id", controller.GetFood())
	routes.POST("/foods", controller.CreateFood())
	routes.PATCH("/foods/:id", controller.UpdateFood())
	routes.DELETE("/foods/:id", controller.DeleteFood())
}
