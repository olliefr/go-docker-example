package main

import (
	"fmt"
	"net/http"
	"os"

	"context"
	"database/sql"
	"log"

	"github.com/cockroachdb/cockroach-go/crdb"
	_ "github.com/lib/pq"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Entry is a single key-value pair
type Entry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func main() {

	// Read the connection parameters from environment variables,
	// following the Twelve-Factor App philosophy: https://12factor.net/
	username := os.Getenv("PGUSER")
	password := os.Getenv("PGPASSWORD")
	hostname := os.Getenv("PGHOST")
	port := os.Getenv("PGPORT")
	database := os.Getenv("PGDATABASE")

	// Build the connection string
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		username, password, hostname, port, database)

	// Initialise the store
	db, err := sql.Open("postgres", connString)
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	defer db.Close()

	// Create the table to store key-value pairs
	if _, err := db.Exec(
		"CREATE TABLE IF NOT EXISTS store (key STRING PRIMARY KEY, value STRING)"); err != nil {
		log.Fatal(err)
	}

	// Initialise the router
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Configure the routes
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, struct{ Status string }{Status: "OK"})
	})
	e.GET("/list", func(c echo.Context) error {
		return listEntries(db, c)
	})
	e.POST("/add", func(c echo.Context) error {
		return addEntry(db, c)
	})

	// Blast off!
	port = os.Getenv("GODOCKER_PORT")
	if port == "" {
		port = "8000"
	}
	e.Logger.Fatal(e.Start(":" + port))
}

// listEntries returns a full copy of the store in JSON format to the client.
func listEntries(db *sql.DB, c echo.Context) error {

	rows, err := db.Query("SELECT key, value FROM store")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Temporary storage for the values read from the DB
	store := map[string]string{}

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			log.Fatal(err)
		}
		store[key] = value
	}

	return c.JSON(http.StatusOK, store)
}

// addEntry adds a new entry to the key-value store
// and returns the value in JSON format to the client.
func addEntry(db *sql.DB, c echo.Context) error {

	e := new(Entry)
	if err := c.Bind(e); err != nil {
		return err
	}

	err := crdb.ExecuteTx(context.Background(), db, nil, func(tx *sql.Tx) error {
		if _, err := tx.Exec(
			"INSERT INTO store (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = excluded.value",
			e.Key, e.Value); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, e)
}
