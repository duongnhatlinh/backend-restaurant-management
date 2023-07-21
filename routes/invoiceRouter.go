package routes

import (
	"github.com/gin-gonic/gin"
	controller "restaurant_management/controllers"
)

func InvoiceRoutes(routes *gin.Engine) {
	routes.GET("/invoices", controller.GetInvoices())
	routes.GET("/invoices/:id", controller.GetInvoice())
	routes.POST("/invoices", controller.CreateInvoice())
	routes.PATCH("/invoices/:id", controller.UpdateInvoice())
	routes.DELETE("/invoices/:id", controller.DeleteInvoice())
}
