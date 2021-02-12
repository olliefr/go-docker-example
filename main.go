package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Entry is a single key-value pair
type Entry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// store is an ordered list of key-value pairs
var store = make([]*Entry, 0)

func main() {

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, struct{ Status string }{Status: "OK"})
	})

	e.GET("/list", listEntries)
	e.POST("/add", addEntry)

	e.Logger.Fatal(e.Start(":8000"))
}

// listEntries returns a full copy of the store in JSON format to the client.
func listEntries(c echo.Context) error {
	return c.JSON(http.StatusOK, store)
}

// addEntry adds a new entry to the key-value store
// and returns the value in JSON format to the client.
func addEntry(c echo.Context) error {
	e := new(Entry)
	if err := c.Bind(e); err != nil {
		return err
	}
	store = append(store, e)
	return c.JSON(http.StatusOK, e)
}
