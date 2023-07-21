package helpers

import (
	"context"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"os"
	"restaurant_management/database"
	"time"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

type SignedDetails struct {
	Email      string
	First_name string
	Last_name  string
	Uid        string
	jwt.StandardClaims
}

var SECRET_KEY = os.Getenv("SECRET_KEY")

func GenerateAllTokens(email string, firstName string, lastName string, uid string) (signedToken, signedRefreshToken string, err error) {
	claims := SignedDetails{
		Email:      email,
		First_name: firstName,
		Last_name:  lastName,
		Uid:        uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	refreshClaims := SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, err
}

func UpdateAllToken(c context.Context, signedToken, signedRefreshToken, userId string) {
	filter := bson.D{{"user_id", userId}}

	var updateObj bson.D
	updateObj = append(updateObj, bson.E{"token", signedToken})
	updateObj = append(updateObj, bson.E{"refresh_token", signedRefreshToken})

	Updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{"updated_at", Updated_at})

	_, err := userCollection.UpdateOne(c, filter, bson.D{{"$set", updateObj}})
	if err != nil {
		log.Panic(err)
		return
	}
	return
}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(signedToken, &SignedDetails{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})

	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = fmt.Sprint("the token is invalid")
		msg = err.Error()
		return nil, msg
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = fmt.Sprint("token is expired")
		return
	}
	return claims, msg
}
