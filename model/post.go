package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// Post use for handling requests from and db storage of posts
type Post struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Title       string             `json:"food,omitempty" bson:"food,omitempty"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	User        string             `json:"user,omitempty" bson:"user,omitempty"`
	PublicURL   string             `json:"publicUrl,omitempty" bson:"publicUrl,omitempty"`
	StorageID   string             `json:"storageId,omitempty" bson:"storageId,omitempty"`
}
