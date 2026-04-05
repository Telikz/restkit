package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"sqlc/db"

	rk "github.com/reststore/restkit"

	_ "github.com/mattn/go-sqlite3"
	"github.com/reststore/restkit/validators/playground"
)

func main() {
	sqlDB, err := sql.Open("sqlite3", "users.db")
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	if _, err := sqlDB.Exec(schema); err != nil {
		log.Fatal(err)
	}

	queries := db.New(sqlDB)

	a := rk.NewApi()
	a.WithVersion("1.0")
	a.WithTitle("User API")
	a.WithDescription("REST API with sqlc")
	a.WithValidator(playground.NewValidator(nil))

	a.WithMiddleware(rk.CORSMiddleware())
	a.WithMiddleware(rk.LoggingMiddleware())

	a.AddGroup(userEndpoints().
		WithMiddleware(rk.DBMiddleware(queries)))

	a.WithSwaggerUI()

	server := http.Server{
		Addr:         ":8080",
		Handler:      a.Mux(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Println("Server on :8080")
	log.Fatal(server.ListenAndServe())
}

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`
