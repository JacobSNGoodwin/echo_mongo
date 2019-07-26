package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Maxbrain0/echo_mongo/model"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// setup mongodB client
	fmt.Println("Establishing connection to MongoDB...")
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))

	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to MongoDB!")

	// setup echo instance and routes
	e := echo.New()
	e.GET("/", helloWorld)
	e.POST("/tacos", postTaco)

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

func helloWorld(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func postTaco(c echo.Context) error {
	meat := c.FormValue("meat")
	description := c.FormValue("description")

	t := &model.Taco{
		Meat:        meat,
		Description: description,
	}

	return c.JSON(http.StatusOK, t)
}
