package controller

import (
	"context"
	"io"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/Maxbrain0/echo_mongo/model"
	"github.com/Maxbrain0/echo_mongo/util"
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

	// before doing transferring files and such, make sure the user is in the database
	// cancel context after time out of if erros
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // use this context for all operations
	defer cancel()

	// get active userID as Object ID
	currentUserID, convErr := primitive.ObjectIDFromHex(util.GetUID(c))

	if convErr != nil {
		cancel()
		return echo.NewHTTPError(http.StatusInternalServerError, "Problem storing data")
	}

	count, err := posts.UserCollection.CountDocuments(ctx, bson.M{"_id": currentUserID})

	if count < 1 || err != nil {
		// need to think about this status code
		cancel()
		return echo.NewHTTPError(http.StatusBadRequest, "User doesn't exist")
	}

	// get request values
	title := c.FormValue("title")
	description := c.FormValue("description")
	image, err := c.FormFile("image")

	if err != nil {
		cancel()
		return err
	}

	// Check to make sure we have an image an limit the file sizee
	mimeTypes := image.Header["Content-Type"]
	if !util.ContainsImage(mimeTypes) {
		cancel()
		return echo.NewHTTPError(http.StatusUnsupportedMediaType, "Image must be of the following file type: jpeg, gif, png, svg, or webp")
	}

	// set a limit on the file size of 10 MB... maybe should be less
	if image.Size > (1024 * 124 * 10) {
		cancel()
		return echo.NewHTTPError(http.StatusRequestEntityTooLarge, "We currently limit the size of image files to 10 Megabytes")
	}

	// open file and send to GC storage
	f, err := image.Open()
	if err != nil {
		cancel()
		return echo.NewHTTPError(http.StatusInternalServerError, "Problem uploading the provided image file")
	}

	defer f.Close()

	// create unique id for file
	storageID := uuid.New().String() + "-" + image.Filename

	o := posts.StorageClient.Bucket("echo-mongo-foodie").Object(storageID)

	wc := o.NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		cancel()
		return echo.NewHTTPError(http.StatusInternalServerError, "Problem uploading the provided image file")
	}
	if err := wc.Close(); err != nil {
		cancel()
		return echo.NewHTTPError(http.StatusInternalServerError, "Problem uploading the provided image file")
	}

	// create url
	url := "https://storage.googleapis.com/echo-mongo-foodie/" + storageID

	// store Post in posts collection, and then add post's storageID to users Posts List
	d := bson.M{"title": title, "description": description, "publicUrl": url, "storageId": storageID, "user": util.GetUserName(c)}
	result, insErr := posts.PostCollection.InsertOne(ctx, d)

	if insErr != nil {
		cancel()
		return echo.NewHTTPError(http.StatusInternalServerError, "Problem storing data")
	}

	// store result id in user's posts array
	oid := result.InsertedID.(primitive.ObjectID)

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

// GetUserPosts extracts the user ID from a json web-token, and returns a list of that user's posts
func (posts *Posts) GetUserPosts(c echo.Context) error {
	// first get the current user from jwt middleware
	uid, err := primitive.ObjectIDFromHex(util.GetUID(c)) // as objectID

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not get user credential")
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	userResp := &model.User{}
	err = posts.UserCollection.FindOne(dbCtx, bson.M{"_id": uid}).Decode(userResp)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No user found. Please login")
	}

	c.JSON(http.StatusOK, userResp)

	return nil
}
