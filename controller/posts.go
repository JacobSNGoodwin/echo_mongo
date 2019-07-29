package controller

import (
	"net/http"

	"github.com/Maxbrain0/echo_mongo/model"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

// Posts holds reference to a database collection and is the receiver of various
// endpoint controllers which will need mongoDB collection access
type Posts struct {
	userCollection *mongo.Collection
	postCollection *mongo.Collection
}

// CreatePost creates (duh) a post for the current user (set in context from jwt middleware)
func (posts *Posts) CreatePost(c echo.Context) error {
	// Key receives an interface, make sure to use type assertion to jwt.Token
	// like wise, Claims is of type (jwt.MapClaims)... Oh delightful type assertion!
	// uid := getUID(c)

	food := c.FormValue("food")
	description := c.FormValue("description")
	image, err := c.FormFile("image")

	if err != nil {
		return err
	}

	response := &model.Post{
		Food:        food,
		Description: description,
		FileURI:     image.Filename,
	}

	return c.JSON(http.StatusOK, response)
}

// utility functions for getting data from cookie - may want to abstrect to util folder of some sort
func getUID(c echo.Context) string {
	return c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["userId"].(string)
}

func getUserName(c echo.Context) string {
	return c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["userName"].(string)
}
