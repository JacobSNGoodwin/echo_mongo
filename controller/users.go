package controller

import (
	"net/http"

	"github.com/Maxbrain0/echo_mongo/model"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

// Users holds reference to a database collection and is the receiver of various
// endpoint controllers which will need mongoDB collection access
type Users struct {
	C *mongo.Collection
}

// CreateUser creates a user in mongo dB and returns a response on success
func (users *Users) CreateUser(c echo.Context) error {
	u := new(model.User)

	if err := c.Bind(u); err != nil {
		return err
	}

	println("User created!")
	return c.JSON(http.StatusCreated, u)
}
