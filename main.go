package main

import (
	"net/http"

	"github.com/Maxbrain0/echo_mongo/model"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.GET("/", helloWorld)
	e.POST("/save", save)
	e.Logger.Fatal(e.Start(":1323"))
}

func helloWorld(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func save(c echo.Context) error {
	meat := c.FormValue("meat")
	description := c.FormValue("description")

	t := &model.Taco{
		Meat:        meat,
		Description: description,
	}

	return c.JSON(http.StatusOK, t)
}
