package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"restaurant_management/database"
	"restaurant_management/models"
	"time"
)

type InvoiceViewFormat struct {
	Invoice_id       string
	Payment_method   string
	Order_id         string
	Payment_status   *string
	Payment_due      interface{}
	Table_number     interface{}
	Payment_due_date time.Time
	Order_details    interface{}
}

var invoiceCollection *mongo.Collection = database.OpenCollection(database.Client, "invoice")

func GetInvoices() gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := invoiceCollection.Find(c, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while listing invoice items"})
		}

		var allInvoices []bson.M
		if err = result.All(c, &allInvoices); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allInvoices)
	}
}

func GetInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		invoiceId := c.Param("id")
		filter := bson.M{"invoice_id": invoiceId}

		var invoice models.Invoice
		if err := invoiceCollection.FindOne(c, filter).Decode(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		allOrderItems, err := ItemsByOrder(invoice.Order_id, c)
		if err != nil {
			msg := fmt.Sprint("error occurred while get items by order")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		var invoiceView InvoiceViewFormat
		invoiceView.Invoice_id = invoice.Invoice_id
		invoiceView.Order_id = invoice.Order_id
		invoiceView.Payment_method = "null"
		if invoice.Payment_method != nil {
			invoiceView.Payment_method = *invoice.Payment_method
		}
		invoiceView.Payment_status = invoice.Payment_status
		invoiceView.Payment_due_date = invoice.Payment_due_date
		invoiceView.Payment_due = allOrderItems[0]["payment_due"]
		invoiceView.Table_number = allOrderItems[0]["table_number"]
		invoiceView.Order_details = allOrderItems[0]["order_items"]

		c.JSON(http.StatusOK, invoiceView)
	}
}

func CreateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var invoice models.Invoice
		if err := c.ShouldBind(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := orderCollection.FindOne(c, bson.M{"order_id": invoice.Order_id}).Err(); err != nil {
			msg := fmt.Sprintf("order not found")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		status := "PENDING"
		if invoice.Payment_status == nil {
			invoice.Payment_status = &status
		}

		invoice.ID = primitive.NewObjectID()
		invoice.Invoice_id = invoice.ID.Hex()
		invoice.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		invoice.Payment_due_date, _ = time.Parse(time.RFC3339, time.Now().AddDate(0, 0, 1).Format(time.RFC3339))

		validationErr := validate.Struct(invoice)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		result, insertErr := invoiceCollection.InsertOne(c, invoice)
		if insertErr != nil {
			msg := fmt.Sprintf("invoice item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func UpdateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		invoiceId := c.Param("id")
		filter := bson.M{"invoice_id": invoiceId}

		var invoice models.Invoice
		if err := c.ShouldBind(&invoice); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var updateObj primitive.D

		if invoice.Payment_method != nil {
			updateObj = append(updateObj, bson.E{"payment_method", invoice.Payment_method})
		}

		if invoice.Payment_status != nil {
			updateObj = append(updateObj, bson.E{"payment_status", invoice.Payment_status})
		}

		invoice.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", invoice.Updated_at})

		result, updateErr := invoiceCollection.UpdateOne(c, filter, updateObj)
		if updateErr != nil {
			msg := fmt.Sprint("invoice update failed")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func DeleteInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var invoiceId = c.Param("id")
		filter := bson.D{{"invoice_id", invoiceId}}

		err := invoiceCollection.FindOne(c, filter).Err()
		if err != nil {
			msg := fmt.Sprint("invoice not found")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		result, deleteErr := invoiceCollection.DeleteOne(c, filter)
		if deleteErr != nil {
			msg := fmt.Sprint("error occurred while delete invoice")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
