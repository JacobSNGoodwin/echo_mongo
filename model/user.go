package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// User contains data for tracking users
type User struct {
	ID       primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	UserName string               `json:"userName" xml:"userName" form:"userName" bson:"userName"`
	Email    string               `json:"email,omitempty" xml:"email,omitempty" form:"email,omitempty" bson:"email,omitempty"`
	Password string               `json:"password,omitempty" xml:"password,omitempty" form:"password,omitempty" bson:"password,omitempty"`
	Posts    []primitive.ObjectID `json:"posts,omitempty" xml:"posts,omitempty" form:"posts,omitempty" bson:"posts,omitempty"`
}
