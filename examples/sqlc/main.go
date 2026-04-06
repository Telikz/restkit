package main

import (
	"database/sql"
	"log"
	"net/http"

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
	a.WithMiddleware(rk.LoggingMiddleware())   // Built-in logging middleware
	a.WithValidator(playground.NewValidator()) // Go-playground style validaiton

	a.AddGroup(rk.NewGroup("/users").
		WithTitle("User Management").
		WithEndpoints(
			getUserEndpoint(queries),
			listUsersEndpoint(queries),
			createUserEndpoint(queries),
			updateUserEndpoint(queries),
			deleteUserEndpoint(queries),
			searchUsersEndpoint(queries),
		),
	)

	a.WithSwaggerUI() // Serve Swagger UI at /swagger

	log.Println("Server on :8080")
	if err := http.ListenAndServe(":8080", a.Mux()); err != nil {
		log.Fatal(err)
	}
}

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`
