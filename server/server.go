package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/Maxbrain0/echo_mongo/model/model"
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
	name := c.FormValue("name")
	email := c.FormValue("email")

	t := &Taco{
		Type:        "Carnitas",
		Description: "Succulent slow cooked pork that is subsequently fried! :)",
	}

	return c.String(http.StatusOK, "name:"+name+", email:"+email)
}
