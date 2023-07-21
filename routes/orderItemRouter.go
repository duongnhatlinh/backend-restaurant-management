package routes

import (
	"github.com/gin-gonic/gin"
	controller "restaurant_management/controllers"
)

func OrderItemRoutes(routes *gin.Engine) {
	routes.GET("/orderItems", controller.GetOrderItems())
	routes.GET("/orderItems/:id", controller.GetOrderItem())
	routes.GET("/orderItems-order/:id", controller.GetOrderItemsByOrder())
	routes.POST("/orderItems", controller.CreateOrderItem())
	routes.PATCH("/orderItems/:id", controller.UpdateOrderItem())
	routes.DELETE("/orderItems/:id", controller.DeleteOrderItem())
}
