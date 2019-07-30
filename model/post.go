package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// Post use for handling requests from and db storage of posts
type Post struct {
	ID          primitive.ObjectID `json:"id" form:"id" query:"id" bson:"_id"`
	Title       string             `json:"title,omitempty" form:"title,omitempty" query:"title,omitempty" bson:"title,omitempty"`
	Description string             `json:"description,omitempty" form:"description,omitempty" query:"description,omitempty" bson:"description,omitempty"`
	User        string             `json:"user,omitempty" form:"user,omitempty" query:"user,omitempty" bson:"user,omitempty"`
	PublicURL   string             `json:"publicUrl,omitempty" form:"publicUrl,omitempty" query:"publicUrl,omitempty" bson:"publicUrl,omitempty"`
	StorageID   string             `json:"storageId,omitempty" form:"storageId,omitempty" query:"storageId,omitempty" bson:"storageId,omitempty"`
}

// PostList will be used for responses retrieving lists of posts
type PostList struct {
	Posts []*Post `json:"posts" bson:"posts" query:"posts"`
	Total int64   `json:"total" bson:"total" query:"total"`
	Limit int64   `json:"limit" bson:"limit" query:"limit"`
	Skip  int64   `json:"skip" bson:"skip" query:"skip"`
}
