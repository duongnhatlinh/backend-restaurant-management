package routes

import (
	"github.com/gin-gonic/gin"
	controller "restaurant_management/controllers"
)

func MenuRoutes(routes *gin.Engine) {
	routes.GET("/menus", controller.GetMenus())
	routes.GET("/menus/:id", controller.GetMenu())
	routes.POST("/menus", controller.CreateMenu())
	routes.PATCH("/menus/:id", controller.UpdateMenu())
	routes.DELETE("/menus/:id", controller.DeleteMenu())
}
