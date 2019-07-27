package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// User contains data for tracking users
type User struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserName string             `json:"userName" bson:"userName"`
	Email    string             `json:"email,omitempty" bson:"email,omitempty"`
	Password string             `json:"password,omitempty" bson:"password, omitempty"`
}
