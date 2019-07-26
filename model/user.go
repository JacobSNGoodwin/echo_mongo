package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// User contains data for tracking users
type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	UserName string             `json:"userName"`
	Email    string             `json:"email"`
	Password string             `json:"password"`
}
