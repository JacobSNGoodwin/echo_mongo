package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// Post use for handling requests from and db storage of posts
type Post struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Food        string             `json:"food" bson:"food"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	FileURI     string             `json:"fileUri,omitempty" bson:"fileUri,omitempty"`
}
