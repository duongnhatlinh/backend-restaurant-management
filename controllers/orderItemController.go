package controllers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"restaurant_management/database"
	"restaurant_management/models"
	"time"
)

type OrderItemPack struct {
	Table_id    *string            `json:"table_id" validate:"required"`
	Order_items []models.OrderItem `json:"order_items"  validate:"required"`
}

var orderItemCollection *mongo.Collection = database.OpenCollection(database.Client, "orderItem")

func GetOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		filter := bson.D{{}}
		cursor, findErr := orderItemCollection.Find(c, filter)
		if findErr != nil {
			msg := fmt.Sprint("error occurred while listing ordered items")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		var orderItems []bson.M
		err := cursor.All(c, &orderItems)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, orderItems)
	}
}

func GetOrderItemsByOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var orderId = c.Param("id")

		allOrderItems, err := ItemsByOrder(orderId, c)

		if err != nil {
			msg := fmt.Sprint("error occurred while listing order items by order ID")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		c.JSON(http.StatusOK, allOrderItems)
	}
}

func GetOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderItemId := c.Param("id")
		filter := bson.D{{"order_item_id", orderItemId}}

		var orderItem models.OrderItem

		if err := orderItemCollection.FindOne(c, filter).Decode(&orderItem); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while fetching the order item"})
			return
		}

		c.JSON(http.StatusOK, orderItem)
	}
}

func CreateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var orderItemPack OrderItemPack
		var order models.Order

		if err := c.BindJSON(&orderItemPack); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		order.Order_date, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if orderItemPack.Table_id == nil {
			msg := fmt.Sprint("Error select table")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}
		order.Table_id = orderItemPack.Table_id

		orderId, err := CreateOrderForOrderItem(c, order)
		if err != nil {
			msg := fmt.Sprint("Err create order for order item")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		orderItemsToBeInserted := []interface{}{}

		for _, orderItem := range orderItemPack.Order_items {
			orderItem.Order_id = orderId
			validationErr := validate.Struct(orderItem)
			if validationErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
				return
			}

			num := toFixed(*orderItem.Unit_price, 2)
			orderItem.Unit_price = &num
			orderItem.ID = primitive.NewObjectID()
			orderItem.Order_item_id = orderItem.ID.Hex()
			orderItem.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

			orderItemsToBeInserted = append(orderItemsToBeInserted, orderItem)
		}

		result, insertOrderItem := orderItemCollection.InsertMany(c, orderItemsToBeInserted)
		if insertOrderItem != nil {
			msg := fmt.Sprint("order items insert failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func UpdateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderItemId := c.Param("id")
		filter := bson.M{"order_item_id": orderItemId}

		var orderItem models.OrderItem

		if err := c.ShouldBind(&orderItem); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var updateObj primitive.D

		if orderItem.Quantity != nil {
			updateObj = append(updateObj, bson.E{"quantity", orderItem.Quantity})
		}

		if orderItem.Unit_price != nil {
			num := toFixed(*orderItem.Unit_price, 2)
			orderItem.Unit_price = &num
			updateObj = append(updateObj, bson.E{"unit_price", orderItem.Unit_price})
		}

		if orderItem.Food_id != nil {
			if err := foodCollection.FindOne(c, bson.D{{"food_id", orderItem.Food_id}}).Err(); err != nil {
				msg := fmt.Sprintf("food not found")
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
			updateObj = append(updateObj, bson.E{"food_id", orderItem.Food_id})
		}

		orderItem.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", orderItem.Updated_at})

		result, updateErr := orderItemCollection.UpdateOne(c, filter, bson.D{{"$set", updateObj}})
		if updateErr != nil {
			msg := fmt.Sprint("order item update failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func DeleteOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var orderItemId = c.Param("id")
		filter := bson.D{{"order_item_id", orderItemId}}

		err := orderItemCollection.FindOne(c, filter).Err()
		if err != nil {
			msg := fmt.Sprint("order item not found")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		result, deleteErr := orderItemCollection.DeleteOne(c, filter)
		if deleteErr != nil {
			msg := fmt.Sprint("error occurred while delete order item")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func ItemsByOrder(OderId string, ctx context.Context) (orderItems []primitive.M, err error) {
	matchStage := bson.D{{"$match", bson.D{{"order_id", OderId}}}}
	lookupFoodStage := bson.D{{"$lookup",
		bson.D{{"from", "food"},
			{"localField", "food_id"},
			{"foreignField", "food_id"},
			{"as", "food"},
		}},
	}
	unwindFoodStage := bson.D{{"$unwind", bson.D{
		{"path", "$food"},
		{"preserveNullAndEmptyArrays", true},
	}}}

	lookupOrderStage := bson.D{{"$lookup",
		bson.D{{"from", "order"},
			{"localField", "order_id"},
			{"foreignField", "order_id"},
			{"as", "order"},
		}}}
	unwindOrderStage := bson.D{{"$unwind",
		bson.D{{"path", "$order"},
			{"preserveNullAndEmptyArrays", true},
		}}}

	lookupTableStage := bson.D{{"$lookup",
		bson.D{{"from", "table"},
			{"localField", "order.table_id"},
			{"foreignField", "table_id"},
			{"as", "table"},
		}}}
	unwindTableStage := bson.D{{"$unwind",
		bson.D{{"path", "$table"},
			{"preserveNullAndEmptyArrays", true},
		}}}

	//addFiledStage := bson.D{{"$addFields", bson.D{{"amount", "$food.price"}}}}
	addFiledStage := bson.D{
		{"$addFields", bson.D{
			{"amount", "$food.price"},
			{"food_name", "$food.name"},
			{"food_image", "$food.food_image"},
			{"table_number", "$table.table_number"},
			{"table_id", "$table.table_id"},
			{"order_id", "$order.order_id"},
			{"price", "$food.price"},
			{"quantity", 1},
		}}}

	groupStage := bson.D{{"$group",
		bson.D{{"_id",
			bson.D{{"order_id", "$order_id"},
				{"table_id", "$table_id"},
				{"table_number", "$table_number"},
			}},
			{"payment_due", bson.D{{"$sum", "$amount"}}},
			{"total_count", bson.D{{"$sum", 1}}},
			{"order_items", bson.D{{"$push", "$$ROOT"}}},
		}}}

	projectStage2 := bson.D{{"$project", bson.D{
		{"_id", 0},
		{"payment_due", 1},
		{"total_count", 1},
		{"table_number", "$_id.table_number"},
		{"order_items", 1},
	}}}

	cursor, err := orderItemCollection.Aggregate(ctx, mongo.Pipeline{
		matchStage,
		lookupFoodStage,
		unwindFoodStage,
		lookupOrderStage,
		unwindOrderStage,
		lookupTableStage,
		unwindTableStage,
		addFiledStage,
		groupStage,
		projectStage2,
	})

	if err != nil {
		return nil, err
	}

	if err := cursor.All(ctx, &orderItems); err != nil {
		return nil, err
	}

	return orderItems, nil
}
