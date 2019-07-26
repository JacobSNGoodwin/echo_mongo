package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Maxbrain0/echo_mongo/model"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Users holds reference to a database collection and is the receiver of various
// endpoint controllers which will need mongoDB collection access
type Users struct {
	Collection *mongo.Collection
}

// CreateUser creates a user in mongo dB and returns a response on success
func (user *Users) CreateUser(c echo.Context) error {
	u := new(model.User)

	if err := c.Bind(u); err != nil {
		return err
	}

	// make sure username and password are available
	if len(u.UserName) < 1 || len(u.Password) < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "Please provide a user name and password")
	}

	// attempt to insert into the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := user.Collection.InsertOne(ctx, bson.M{"userName": u.UserName, "password": u.Password, "email": u.Email})

	//
	if err != nil {
		fmt.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not add user")
	}

	oid := res.InsertedID.(primitive.ObjectID)

	fmt.Println(oid)

	response := &model.User{
		ID:       oid,
		UserName: u.UserName,
	}

	return c.JSON(http.StatusCreated, response)
}
