package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// User contains data for tracking users
type User struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserName string             `json:"userName" xml:"userName" form:"userName" bson:"userName"`
	Email    string             `json:"email,omitempty" xml:"email,omitempty" form:"email,omitempty" bson:"email,omitempty"`
	Password string             `json:"password,omitempty" xml:"password,omitempty" form:"password,omitempty" bson:"password, omitempty"`
	Posts    []string           `json:"posts,omitempty" xml:"posts,omitempty" form:"posts,omitempty" bson:"posts, omitempty"`
}
