package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"cloud.google.com/go/storage"
	"github.com/Maxbrain0/echo_mongo/controller"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// global flags set via command line - an example bash script is included for some reasonable settings
var dburi string
var gcconfig string

// global server, controllers, collections, and handle to cloud storage
var e *echo.Echo
var gcClient *storage.Client
var userCollection *mongo.Collection
var postCollection *mongo.Collection
var usersController *controller.Users
var postsController *controller.Posts

// init used to parse flags
func init() {
	flag.StringVar(&dburi, "dburi", "mongodb://root:example@localhost:27017", "The db of the mongo URI. The default URI for a docker container is included.")
	flag.StringVar(&gcconfig, "gcconfig", "", "The path of the json config file for Google Cloud Storage. See https://cloud.google.com/storage/docs/reference/libraries#client-libraries-install-go for more information")

	flag.Parse()

	// set Google cloud environment variable from command line
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", gcconfig)

	// fmt.Println(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	// fmt.Println("Starting on the following db uri: ", dburi)
	// fmt.Println("Using the following Google Cloud Config: ", gcconfig)
}

func main() {
	// setup mongodB client
	fmt.Println("Establishing connection to MongoDB...")
	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))

	if err != nil {
		log.Fatal(err)
	}

	// use this context timeout for both mongo and google cloud storage client
	// cancel if either fails
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)

	if err != nil {
		cancel()
		log.Fatal(err)
	}

	// might want to ping here to really make sure we're connected
	fmt.Println("Successfully connected to MongoDB!")

	// add a userCollection and postCollection
	userCollection = client.Database("foodie").Collection("users")
	postCollection = client.Database("foodie").Collection("posts")

	// Setup client connection to google cloud
	// Sets your Google Cloud Platform project ID.
	fmt.Println("Creating Google Cloud Storage Client")

	// Creates a client.
	gcClient, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("Successfully Created Google Cloud Storage Client")

	// routes are configured below, main more for setup and teardown
	setupRoutes()

	// Provide global connection clients to controllers
	usersController = &controller.Users{Collection: userCollection}
	postsController = &controller.Posts{UserCollection: userCollection, PostCollection: postCollection, StorageClient: gcClient}

	// allows us to shut down server gracefully
	go func() {
		if err := e.Start(":1323"); err != nil {
			e.Logger.Info("shutting down the server")
		}
	}()

	// Wait for Control C to exit - shut down mongo and server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	// Block until a signal is received
	<-quit

	// shut down echo server
	fmt.Println("Shutting down the echo server...")
	ctxDisconnect, cancelDisconnect := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelDisconnect()
	if err := e.Shutdown(ctxDisconnect); err != nil {
		e.Logger.Fatal(err)
	}
	fmt.Println("Successfully shut down echo server!")

	// shut down mongo db
	fmt.Println("Disconnecting from MongoDB...")

	if err := client.Disconnect(ctxDisconnect); err != nil {
		log.Fatal("Problem shutting down mongodb\n")
	}

	fmt.Println("Succesfully Disconnected from MongoDB")
}

/*
* Setup routes for echo rest api here
 */
func setupRoutes() {
	// jwt middleware config
	config := middleware.JWTConfig{
		SigningKey:  []byte("secret"),
		TokenLookup: "cookie:token",
	}
	jwtmw := middleware.JWTWithConfig(config)

	// setup echo instance and routes

	e = echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.POST("/user", usersController.CreateUser)
	e.POST("/login", usersController.Login)

	// Must have authentication to create a post, so apss jwt middleware
	e.POST("/post", postsController.CreatePost, jwtmw)

}
