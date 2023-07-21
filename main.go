package main

import (
	"github.com/gin-gonic/gin"
	"os"
	"restaurant_management/middleware"
	"restaurant_management/routes"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	router := gin.New()
	router.Use(gin.Logger())
	routes.UserRoutes(router)
	router.Use(middleware.Authentication())

	routes.FoodRoutes(router)
	routes.MenuRoutes(router)
	routes.TableRoutes(router)
	routes.OrderItemRoutes(router)
	routes.OrderRoutes(router)
	routes.InvoiceRoutes(router)

	router.Run(":" + port)
}
