package controller

import (
	"net/http"

	"cloud.google.com/go/storage"
	"github.com/Maxbrain0/echo_mongo/model"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

// Posts holds reference to a database collection and is the receiver of various
// endpoint controllers which will need mongoDB collection access
type Posts struct {
	UserCollection *mongo.Collection
	PostCollection *mongo.Collection
	StorageClient  *storage.Client
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

	// Check to make sure we have an image an limit the file sizee
	mimeTypes := image.Header["Content-Type"]
	if !containsImage(mimeTypes) {
		return echo.NewHTTPError(http.StatusUnsupportedMediaType, "Image must be of the following file type: jpeg, gif, png, svg, or webp")
	}

	// set a limit on the file size of 10 MB... maybe should be less
	if image.Size > (1024 * 124 * 10) {
		return echo.NewHTTPError(http.StatusRequestEntityTooLarge, "We currently limit the size of image files to 10 Megabytes")
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

// function to make sure we have a mime-type of an image (in case of multiple mime-types, which I'm not sure actually happens often)
func containsImage(s []string) bool {
	for _, a := range s {
		if a == "image/jpeg" || a == "image/gif" || a == "image/png" || a == "image/svg+xml" || a == "image/webp" {
			return true
		}
	}
	return false
}
