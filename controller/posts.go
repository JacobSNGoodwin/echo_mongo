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
	Collection *mongo.Collection
}

// CreatePost creates (duh) a post for the current user (set in context from jwt middleware)
func (posts *Posts) CreatePost(c echo.Context) error {
	p := new(model.Post)

	if err := c.Bind(p); err != nil {
		return err
	}

	// Key receives an interface, make sure to use type assertion to jwt.Token
	// like wise, Claims is of type (jwt.MapClaims)... Oh delightful type assertion!
	claimMap := c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)

	return c.String(http.StatusOK, "Thanks,"+claimMap["userName"].(string))
}
