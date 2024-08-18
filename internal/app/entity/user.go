package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id" json:"id"`
	Name     string             `bson:"name" json:"name"`
	Password string             `bson:"password" json:"password"`
}
