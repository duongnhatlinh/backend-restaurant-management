package routes

import (
	"github.com/gin-gonic/gin"
	controller "restaurant_management/controllers"
)

func OrderRoutes(routes *gin.Engine) {
	routes.GET("/orders", controller.GetOrders())
	routes.GET("/orders/:id", controller.GetOrder())
	routes.POST("/orders", controller.CreateOrder())
	routes.PATCH("/orders/:id", controller.UpdateOrder())
	routes.DELETE("/orders/:id", controller.DeleteOrder())
}
