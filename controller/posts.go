package controller

import (
	"context"
	"io"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/Maxbrain0/echo_mongo/model"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	title := c.FormValue("title")
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

	// open file and send to GC storage
	f, err := image.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Problem uploading the provided image file")
	}

	defer f.Close()

	// create unique id for file
	storageID := uuid.New().String() + "-" + image.Filename

	// consider with timeout... need to determine reasonable time for this operation
	ctx := context.Background()

	o := posts.StorageClient.Bucket("echo-mongo-foodie").Object(storageID)

	wc := o.NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Problem uploading the provided image file")
	}
	if err := wc.Close(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Problem uploading the provided image file")
	}

	// create url
	url := "https://storage.googleapis.com/echo-mongo-foodie/" + storageID

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// store Post in posts collection, and then add post's storageID to users Posts List
	d := bson.M{"title": title, "description": description, "publicUrl": url, "storageId": storageID, "user": getUserName(c)}
	result, insErr := posts.PostCollection.InsertOne(ctx, d)

	if insErr != nil {
		cancel()
		return echo.NewHTTPError(http.StatusInternalServerError, "Problem storing data")
	}

	// store result id in user's posts array
	oid := result.InsertedID.(primitive.ObjectID)

	// get active userID as Object ID
	currentUserID, convErr := primitive.ObjectIDFromHex(getUID(c))

	if convErr != nil {
		cancel()
		return echo.NewHTTPError(http.StatusInternalServerError, "Problem storing data")
	}

	// update record of currently authenticated user... add to this user's posts array
	updateErr := posts.UserCollection.FindOneAndUpdate(ctx, bson.M{"_id": currentUserID}, bson.M{"$addToSet": bson.M{"posts": oid}}).Err()

	if updateErr != nil {
		cancel()
		return echo.NewHTTPError(http.StatusInternalServerError, "Problem storing data")
	}

	response := &model.Post{
		ID: oid,
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
