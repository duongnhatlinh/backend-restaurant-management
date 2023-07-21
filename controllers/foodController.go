package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"math"
	"net/http"
	"restaurant_management/database"
	"restaurant_management/models"
	"strconv"
	"time"
)

var validate = validator.New()
var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

// GetFoods listing food items
func GetFoods() gin.HandlerFunc {
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
			}}}
		projectStage := bson.D{
			{
				"$project", bson.D{
					{"_id", 0},
					{"total_count", 1},
					{"food_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
				}},
		}

		result, err := foodCollection.Aggregate(c, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while listing food items"})
			return
		}

		var allFoods []bson.M
		if err = result.All(c, &allFoods); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allFoods)
	}
}

// GetFood fetching the food item
func GetFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		foodId := c.Param("id")
		var food models.Food

		if err := foodCollection.FindOne(c, bson.D{{"food_id", foodId}}).Decode(&food); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while fetching the food item"})
			return
		}

		c.JSON(http.StatusOK, food)
	}
}

// CreateFood create new food
func CreateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var food models.Food

		if err := c.ShouldBind(&food); err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		validationErr := validate.Struct(food)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		if err := menuCollection.FindOne(c, bson.M{"menu_id": food.Menu_id}).Err(); err != nil {
			msg := fmt.Sprintf("menu not found")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		food.ID = primitive.NewObjectID()
		food.Food_id = food.ID.Hex()
		food.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		num := toFixed(*food.Price, 2)
		food.Price = &num

		result, insertErr := foodCollection.InsertOne(c, food)
		if insertErr != nil {
			msg := fmt.Sprintf("Food item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// toFixed give a fixed number
func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return math.Round(num*output) / output
}

func UpdateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		foodId := c.Param("id")
		filter := bson.M{"food_id": foodId}

		var food models.Food

		if err := c.ShouldBind(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var updateObj primitive.D

		if food.Name != nil {
			updateObj = append(updateObj, bson.E{"name", food.Name})
		}

		if food.Price != nil {
			num := toFixed(*food.Price, 2)
			food.Price = &num
			updateObj = append(updateObj, bson.E{"price", food.Price})
		}

		if food.Food_image != nil {
			updateObj = append(updateObj, bson.E{"food_image", food.Food_image})
		}

		if food.Menu_id != nil {
			if err := menuCollection.FindOne(c, filter).Err(); err != nil {
				msg := fmt.Sprintf("menu not found")
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
			updateObj = append(updateObj, bson.E{"menu_id", food.Menu_id})
		}

		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", food.Updated_at})

		result, updateErr := foodCollection.UpdateOne(c, filter, bson.D{{"$set", updateObj}})
		if updateErr != nil {
			msg := fmt.Sprint("food item update failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func DeleteFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var foodId = c.Param("id")
		filter := bson.D{{"food_id", foodId}}

		err := foodCollection.FindOne(c, filter).Err()
		if err != nil {
			msg := fmt.Sprint("food item not found")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		result, deleteErr := foodCollection.DeleteOne(c, filter)
		if deleteErr != nil {
			msg := fmt.Sprint("error occurred while delete food item")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
