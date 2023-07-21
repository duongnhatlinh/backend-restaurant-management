package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Menu struct {
	ID         primitive.ObjectID `bson:"_id"`
	Name       string             `json:"name" validate:"required"`
	Category   string             `json:"category" validate:"required"`
	Start_date *time.Time         `json:"start_date"`
	End_date   *time.Time         `json:"end_date"`
	Created_at time.Time          `json:"created_at"`
	Updated_at time.Time          `json:"update_at"`
	Menu_id    string             `json:"menu_id"`
}
