package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Maxbrain0/echo_mongo/controller"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// setup mongodB client
	fmt.Println("Establishing connection to MongoDB...")
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://root:example@localhost:27017"))

	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)

	if err != nil {
		log.Fatal(err)
	}

	// might want to ping here to really make sure we're connected
	fmt.Println("Successfully connected to MongoDB!")

	// setup a users collection
	userCollection := client.Database("foodie").Collection("users")

	usersController := &controller.Users{Collection: userCollection}

	// setup echo instance and routes
	// we wrap functions that need to pass the collection
	e := echo.New()
	e.POST("/user", usersController.CreateUser)
	e.POST("/login", usersController.Login)

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
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	fmt.Println("Successfully shut down echo server!")

	// shut down mongo db
	fmt.Println("Disconnecting from MongoDB...")

	if err := client.Disconnect(ctx); err != nil {
		log.Fatal("Problem shutting down mongodb")
	}

	fmt.Println("Succesfully Disconnect from MongoDB")
}
