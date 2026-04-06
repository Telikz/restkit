package main

import (
	"context"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	rk "github.com/reststore/restkit"
	restecho "github.com/reststore/restkit/adapters/echo"
)

func main() {
	e := echo.New()

	api := rk.NewApi().
		WithVersion("1.0.0").
		WithTitle("Echo + RestKit Example").
		WithSwaggerUI("/docs")

	api.AddGroup(userGroup())
	api.AddEndpoint(ping())
	restecho.RegisterRoutes(e, api)

	echoRouter := echo.New()
	echoRouter.GET("/native/users", echoListUsers)
	echoRouter.GET("/native/users/:id", echoGetUser)
	echoRouter.POST("/native/users", echoCreateUser)

	meta := []rk.RouteMeta{
		{
			Method: "GET",
			Path:   "/native/users",
			Info: rk.RouteInfo{
				Summary:      "List users",
				ResponseType: []User{},
			},
		},
		{
			Method: "GET",
			Path:   "/native/users/:id",
			Info: rk.RouteInfo{
				Summary:      "Get user",
				ResponseType: User{},
			},
		},
		{
			Method: "POST",
			Path:   "/native/users",
			Info: rk.RouteInfo{
				Summary:      "Create user",
				RequestType:  CreateUserReq{},
				ResponseType: User{},
			},
		},
	}

	_ = restecho.Mount(api, "/api/v1", echoRouter, meta)

	log.Println("Server on :8080")
	log.Fatal(http.ListenAndServe(":8080", api.Mux()))
}

type CreateUserReq struct {
	Name  string `json:"name"  validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func userGroup() *rk.Group {
	return rk.NewGroup("/users").
		WithEndpoints(
			createUser(),
			getUser(),
			listUsers(),
		)
}

func createUser() *rk.Endpoint[CreateUserReq, User] {
	return rk.NewEndpoint[CreateUserReq, User]().
		WithPath("/").
		WithMethod("POST").
		WithHandler(func(ctx context.Context, req CreateUserReq) (User, error) {
			return User{ID: 1, Name: req.Name, Email: req.Email}, nil
		})
}

func getUser() *rk.Endpoint[rk.NoRequest, User] {
	return rk.NewEndpoint[rk.NoRequest, User]().
		WithPath("/:id").
		WithMethod("GET").
		WithHandler(func(ctx context.Context, _ rk.NoRequest) (User, error) {
			return User{ID: 1, Name: "John", Email: "john@example.com"}, nil
		})
}

func listUsers() *rk.Endpoint[rk.NoRequest, []User] {
	return rk.NewEndpoint[rk.NoRequest, []User]().
		WithPath("/").
		WithMethod("GET").
		WithHandler(func(ctx context.Context, _ rk.NoRequest) ([]User, error) {
			return []User{{ID: 1, Name: "John", Email: "john@example.com"}}, nil
		})
}

func ping() *rk.Endpoint[rk.NoRequest, Pong] {
	return rk.NewEndpoint[rk.NoRequest, Pong]().
		WithPath("/ping").
		WithMethod("GET").
		WithHandler(func(ctx context.Context, _ rk.NoRequest) (Pong, error) {
			return Pong{Message: "pong"}, nil
		})
}

type Pong struct {
	Message string `json:"message"`
}

func echoListUsers(c echo.Context) error {
	return c.JSON(200, []User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
	})
}

func echoGetUser(c echo.Context) error {
	id := c.Param("id")
	return c.JSON(200, User{ID: 1, Name: "Alice (" + id + ")"})
}

func echoCreateUser(c echo.Context) error {
	var req CreateUserReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, echo.Map{"error": err.Error()})
	}
	return c.JSON(201, User{ID: 3, Name: req.Name, Email: req.Email})
}
