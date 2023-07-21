package controllers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"restaurant_management/database"
	"restaurant_management/models"
	"strconv"
	"time"
)

var orderCollection *mongo.Collection = database.OpenCollection(database.Client, "order")

func GetOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if recordPerPage < 1 {
			recordPerPage = 2
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage

		matchStage := bson.D{{"$match", bson.D{{}}}}
		groupStage := bson.D{
			{"$group", bson.D{
				{"_id", bson.D{{"_id", "null"}}},
				{"total_count", bson.D{{"$sum", 1}}},
				{"data", bson.D{{"$push", "$$ROOT"}}},
			}},
		}
		projectStage := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"order_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
			}},
		}

		result, err := orderCollection.Aggregate(c, mongo.Pipeline{
			matchStage,
			groupStage,
			projectStage,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while listing order items"})
			return
		}

		var allOrders []bson.M
		if err := result.All(c, &allOrders); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allOrders)
	}
}

func GetOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderId := c.Param("id")
		filter := bson.M{"order_id": orderId}

		var order models.Order
		err := orderCollection.FindOne(c, filter).Decode(&order)
		if err != nil {
			msg := fmt.Sprint("order not found")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, order)
	}
}

func CreateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var order models.Order

		if err := c.ShouldBind(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validateErr := validate.Struct(order)
		if validateErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validateErr.Error()})
			return
		}

		if !checkOrderDate(order.Order_date) {
			msg := fmt.Sprint("kindly retype the time")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		if err := tableCollection.FindOne(c, bson.D{{"table_id", order.Table_id}}).Err(); err != nil {
			msg := fmt.Sprint("table not found")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		order.ID = primitive.NewObjectID()
		order.Order_id = order.ID.Hex()
		order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		result, insertErr := orderCollection.InsertOne(c, order)
		if insertErr != nil {
			msg := fmt.Sprint("order was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func checkOrderDate(date time.Time) bool {
	return date.After(time.Now())
}
func UpdateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderId := c.Param("id")
		filter := bson.M{"order_id": orderId}

		var order models.Order
		if err := c.ShouldBind(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var updateObj primitive.D

		if order.Table_id != nil {
			errTable := tableCollection.FindOne(c, bson.D{{"table_id", order.Table_id}}).Err()
			if errTable != nil {
				msg := fmt.Sprint("table not found")
				c.JSON(http.StatusBadRequest, gin.H{"error": msg})
				return
			}
			updateObj = append(updateObj, bson.E{"table_id", order.Table_id})
		}

		if !checkOrderDate(order.Order_date) {
			msg := fmt.Sprint("kindly retype the time")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}
		updateObj = append(updateObj, bson.E{"order_date", order.Order_date})

		order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", order.Updated_at})

		result, updateErr := orderCollection.UpdateOne(c, filter, bson.D{{"$set", updateObj}})
		if updateErr != nil {
			msg := fmt.Sprint("order item update failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func DeleteOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var orderId = c.Param("id")
		filter := bson.D{{"order_id", orderId}}

		err := orderCollection.FindOne(c, filter).Err()
		if err != nil {
			msg := fmt.Sprint("order item not found")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		result, deleteErr := orderCollection.DeleteOne(c, filter)
		if deleteErr != nil {
			msg := fmt.Sprint("error occurred while delete order item")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func CreateOrderForOrderItem(ctx context.Context, order models.Order) (string, error) {
	order.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	order.ID = primitive.NewObjectID()
	order.Order_id = order.ID.Hex()

	_, insertErr := orderCollection.InsertOne(ctx, order)
	if insertErr != nil {
		return "", insertErr
	}

	return order.Order_id, nil
}
