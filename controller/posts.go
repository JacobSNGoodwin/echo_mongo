package controller

import (
	"context"
	"fmt"
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
	"go.mongodb.org/mongo-driver/mongo/options"
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
	uid, err := primitive.ObjectIDFromHex(util.GetUID(c)) // as ObjectID

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not get user credential")
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	// retrieve limit and skip
	params := new(model.PostList)

	if err := c.Bind(params); err != nil {
		return err
	}

	// retrieve user's list of post ObjectID's from UserCollection - need to return total count, too
	userResp := &model.User{}
	err = posts.UserCollection.FindOne(dbCtx, bson.M{"_id": uid}).Decode(userResp)

	if err != nil {
		dbCancel()
		return echo.NewHTTPError(http.StatusBadRequest, "No user found. Please login")
	}

	// get actual post data from PostCollection - default sort, use limit and skip
	findOptions := options.Find()
	findOptions.SetLimit(params.Limit)
	findOptions.SetSkip(params.Skip)

	cursor, err := posts.PostCollection.Find(dbCtx, bson.M{"_id": bson.M{"$in": userResp.Posts}}, findOptions)

	if err != nil {
		dbCancel()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// decode response into a slice of posts
	respPosts := []*model.Post{}

	for cursor.Next(dbCtx) {
		elem := &model.Post{} // type to decode into... dangling preposition! O shame!
		if err := cursor.Decode(elem); err != nil {
			dbCancel()
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		respPosts = append(respPosts, elem)
	}

	if err := cursor.Err(); err != nil {
		dbCancel()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// The final response
	resp := &model.PostList{
		Posts: respPosts,
		Total: int64(len(userResp.Posts)),
		Limit: params.Limit,
		Skip:  params.Skip,
	}

	return c.JSON(http.StatusOK, resp)
}

// GetPosts gets all posts that are public
// Note, currently users don't have public or private post settings, but can in the future
func (posts *Posts) GetPosts(c echo.Context) error {
	// retrieve limit and skip
	params := new(model.PostList)
	if err := c.Bind(params); err != nil {
		return err
	}

	// get actual post data from PostCollection - use limit and skip
	findOptions := options.Find()
	findOptions.SetLimit(params.Limit)
	findOptions.SetSkip(params.Skip)

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	// get totall count on collection using metadata since we're not filtering
	postCount, err := posts.PostCollection.EstimatedDocumentCount(dbCtx)

	// use a find without filters and above FindOptions
	cursor, err := posts.PostCollection.Find(dbCtx, bson.M{}, findOptions)
	if err != nil {
		dbCancel()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err != nil {
		dbCancel()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// decode response into a slice of posts
	respPosts := []*model.Post{}

	for cursor.Next(dbCtx) {
		elem := &model.Post{} // type to decode into... dangling preposition! O shame!
		if err := cursor.Decode(elem); err != nil {
			dbCancel()
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		respPosts = append(respPosts, elem)
	}

	if err := cursor.Err(); err != nil {
		dbCancel()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// The final response
	resp := &model.PostList{
		Posts: respPosts,
		Total: postCount,
		Limit: params.Limit,
		Skip:  params.Skip,
	}

	return c.JSON(http.StatusOK, resp)
}

// DeletePost retrieves the ID of a post from url and deletes it given that the psot belongs
// to the current user stored in the jwt in the context
func (posts *Posts) DeletePost(c echo.Context) error {
	// fetch the PostID and make sure it is in the current user's list
	postID, err := primitive.ObjectIDFromHex(c.Param("id"))

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Could not parse provided id. Please provide a valid post id")
	}

	// check is zero objectID (ie, no id provided by query body or params)
	if postID.IsZero() {
		return echo.NewHTTPError(http.StatusBadRequest, "Please provide the document ID as a query parameter, or in the body as 'id'")
	}

	// get current userID
	uid, err := primitive.ObjectIDFromHex(util.GetUID(c)) // as ObjectID

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not get user credential")
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()

	// try to update document for this user by deleting it from their posts list
	// if there's an error, or this doesn't exist in their posts list, we won't delete
	// any items from the Posts collection
	updateResult, err := posts.UserCollection.UpdateOne(
		dbCtx,
		bson.M{
			"_id": uid,
		},
		bson.M{
			"$pull": bson.M{
				"posts": postID,
			},
		},
	)

	if err != nil {
		dbCancel()
		return echo.NewHTTPError(http.StatusBadRequest, "Could not remove post for current user.")
	}

	// if the post list was not modified
	if updateResult.ModifiedCount == 0 {
		dbCancel()
		return echo.NewHTTPError(http.StatusBadRequest, "Could not remove post for current user.")
	}

	// if we did modify users list, we can delete post from Posts Collection
	deleteResult, err := posts.PostCollection.DeleteOne(dbCtx, bson.M{
		"_id": postID,
	})

	if err != nil || deleteResult.DeletedCount < 1 {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete document")
	}

	return c.JSON(http.StatusOK, bson.M{
		"message":          fmt.Sprintf("Successfully removed post with the following id: %v", postID.Hex()),
		"deletedPostCount": deleteResult.DeletedCount,
	})
}
