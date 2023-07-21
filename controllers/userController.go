package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"restaurant_management/database"
	"restaurant_management/helpers"
	"restaurant_management/models"
	"strconv"
	"time"
)

type UserLogin struct {
	Email    *string `json:"email" validate:"email,required"`
	Password *string `json:"password" validate:"required,min=6"`
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GetUsers() gin.HandlerFunc {
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
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
			}}}

		result, err := userCollection.Aggregate(c, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while listing food users"})
			return
		}

		var allUsers []bson.M
		if err = result.All(c, &allUsers); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allUsers)

	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("id")

		var user models.User

		err := userCollection.FindOne(c, bson.M{"user_id": userId}).Decode(&user)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while get user"})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		countEmail, err := userCollection.CountDocuments(c, bson.D{{"email", user.Email}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while checking for the email"})
			return
		}

		if countEmail > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email  already exist"})
			return
		}

		countPhone, err := userCollection.CountDocuments(c, bson.D{{"phone", user.Phone}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while checking for the phone number"})
			return
		}

		if countPhone > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this phone number  already exist"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		token, refreshToken, _ := helpers.GenerateAllTokens(
			*user.Email,
			*user.First_name,
			*user.Last_name,
			user.User_id,
		)

		user.Token = &token
		user.Refresh_token = &refreshToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(c, user)
		if insertErr != nil {
			msg := fmt.Sprintf("User item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func LogIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		var userLogin UserLogin
		var foundUser models.User

		if err := c.BindJSON(&userLogin); err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		err := userCollection.FindOne(c, bson.M{"email": userLogin.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found, login seems to be incorrect"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*foundUser.Password, *userLogin.Password)
		if passwordIsValid != true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		token, refreshToken, _ := helpers.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, foundUser.User_id)
		helpers.UpdateAllToken(c, token, refreshToken, foundUser.User_id)

		c.JSON(http.StatusOK, foundUser)
	}
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

func VerifyPassword(userPassword, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(userPassword), []byte(providedPassword))
	check := true
	msg := ""
	if err != nil {
		msg = fmt.Sprintf("login or password is incorrect")
		check = false
	}
	return check, msg
}
