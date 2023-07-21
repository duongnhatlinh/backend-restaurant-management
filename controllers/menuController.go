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

var menuCollection *mongo.Collection = database.OpenCollection(database.Client, "menu")

func GetMenus() gin.HandlerFunc {
	return func(c *gin.Context) {
		cursor, err := menuCollection.Find(c, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var allMenus []bson.M
		if err := cursor.All(c, &allMenus); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allMenus)
	}
}

func GetMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		menuId := c.Param("id")
		filter := bson.D{{"menu_id", menuId}}

		var menu models.Menu

		err := menuCollection.FindOne(c, filter).Decode(&menu)
		if err != nil {
			msg := fmt.Sprint("Menu item not found")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, menu)
	}
}

func CreateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		var menu models.Menu

		if err := c.ShouldBind(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if validateErr := validate.Struct(menu); validateErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validateErr.Error()})
			return
		}

		if !inTimeSpan(*menu.Start_date, *menu.End_date) {
			msg := fmt.Sprint("kindly retype the time")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		menu.ID = primitive.NewObjectID()
		menu.Menu_id = menu.ID.Hex()
		menu.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		menu.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		result, insertErr := menuCollection.InsertOne(c, menu)
		if insertErr != nil {
			msg := fmt.Sprint("Menu item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func inTimeSpan(start, end time.Time) bool {
	return start.After(time.Now()) && end.After(start)
}

func UpdateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		menuId := c.Param("id")
		filter := bson.D{{"menu_id", menuId}}

		var menu models.Menu
		if err := c.ShouldBind(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}

		var menuObj primitive.D
		if menu.Start_date != nil && menu.End_date != nil {
			if !inTimeSpan(*menu.Start_date, *menu.End_date) {
				msg := fmt.Sprint("kindly retype the time")
				c.JSON(http.StatusBadRequest, gin.H{"error": msg})
				return
			}
			menuObj = append(menuObj, bson.E{"start_date", menu.Start_date})
			menuObj = append(menuObj, bson.E{"end_date", menu.End_date})
		}

		if menu.Name != "" {
			menuObj = append(menuObj, bson.E{"name", menu.Name})
		}

		if menu.Category != "" {
			menuObj = append(menuObj, bson.E{"category", menu.Category})
		}

		menu.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		menuObj = append(menuObj, bson.E{"updated_at", menu.Updated_at})

		result, updateErr := menuCollection.UpdateOne(c, filter, bson.D{
			{"$set", menuObj},
		})

		if updateErr != nil {
			msg := fmt.Sprint("menu item update failed")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func DeleteMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		var menuId = c.Param("id")
		filter := bson.D{{"menu_id", menuId}}

		err := menuCollection.FindOne(c, filter).Err()
		if err != nil {
			msg := fmt.Sprint("menu item not found")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		result, deleteErr := menuCollection.DeleteOne(c, filter)
		if deleteErr != nil {
			msg := fmt.Sprint("error occurred while delete menu item")
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
