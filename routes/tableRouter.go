package routes

import (
	"github.com/gin-gonic/gin"
	controller "restaurant_management/controllers"
)

func TableRoutes(routes *gin.Engine) {
	routes.GET("/tables", controller.GetTables())
	routes.GET("/tables/:id", controller.GetTable())
	routes.POST("/tables", controller.CreateTable())
	routes.PATCH("/tables/:id", controller.UpdateTable())
	routes.DELETE("/tables/:id", controller.DeleteTable())
}
