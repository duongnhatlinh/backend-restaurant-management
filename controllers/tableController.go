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

var tableCollection *mongo.Collection = database.OpenCollection(database.Client, "table")

func GetTables() gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := tableCollection.Find(c, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while listing table"})
		}

		var allTables []bson.M
		if err = result.All(c, &allTables); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allTables)
	}
}

func GetTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		tableId := c.Param("id")
		var table models.Table

		if err := tableCollection.FindOne(c, bson.D{{"table_id", tableId}}).Decode(&table); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while fetching the table"})
			return
		}

		c.JSON(http.StatusOK, table)
	}
}

func CreateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		var table models.Table

		if err := c.ShouldBind(&table); err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		validationErr := validate.Struct(table)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		table.ID = primitive.NewObjectID()
		table.Table_id = table.ID.Hex()
		table.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		table.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		result, insertErr := tableCollection.InsertOne(c, table)
		if insertErr != nil {
			msg := fmt.Sprintf("table was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func UpdateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		tableId := c.Param("id")
		filter := bson.M{"table_id": tableId}

		var table models.Table
		if err := c.ShouldBind(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var updateObj primitive.D

		if table.Number_of_guests != nil {
			updateObj = append(updateObj, bson.E{"number_of_guests", table.Number_of_guests})
		}

		if table.Table_number != nil {
			updateObj = append(updateObj, bson.E{"table_number", table.Table_number})
		}

		table.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", table.Updated_at})

		result, updateErr := tableCollection.UpdateOne(c, filter, bson.D{{"$set", updateObj}})
		if updateErr != nil {
			msg := fmt.Sprint("table update failed")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func DeleteTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tableId = c.Param("id")
		filter := bson.D{{"table_id", tableId}}

		err := tableCollection.FindOne(c, filter).Err()
		if err != nil {
			msg := fmt.Sprint("table not found")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		result, deleteErr := tableCollection.DeleteOne(c, filter)
		if deleteErr != nil {
			msg := fmt.Sprint("error occurred while delete table")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
