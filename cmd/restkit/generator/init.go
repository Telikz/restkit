package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

func InitProject(moduleName string) error {
	if moduleName == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		moduleName = filepath.Base(wd)
	}

	if _, err := os.Stat("go.mod"); err != nil {
		cmd := exec.Command("go", "mod", "init", moduleName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to initialize go module: %w", err)
		}
	}

	dirs := []string{"endpoints"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	type mainData struct {
		Module string
	}
	if err := generateTemplateFile("main.go", mainTemplate, mainData{Module: moduleName}); err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}

	if err := os.WriteFile("endpoints/ping.go", []byte(pingEndpointTemplate), 0o644); err != nil {
		return fmt.Errorf("failed to create ping endpoint: %w", err)
	}

	cmd := exec.Command("go", "get", "github.com/RestStore/RestKit")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add restkit dependency: %w", err)
	}

	cmd = exec.Command("go", "fmt", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to format code: %w", err)
	}

	cmd = exec.Command("go", "mod", "tidy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to tidy go module: %w", err)
	}

	return nil
}

func generateTemplateFile(filename, tmplStr string, data interface{}) error {
	tmpl, err := template.New("file").Parse(tmplStr)
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

const mainTemplate = `package main

import (
	"log"
	"net/http"

	"{{.Module}}/endpoints"

	"github.com/RestStore/RestKit"
)

func main() {
	api := restkit.NewApi()
	api.WithTitle("My API")
	api.WithVersion("0.0.1")
	api.WithDescription("My API description")

	api.Add(endpoints.Ping())

	api.WithSwaggerUI("/docs")

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", api.Mux()); err != nil {
		log.Fatal(err)
	}
}
`

const pingEndpointTemplate = `package endpoints

import (
	"context"
	"net/http"

	"github.com/RestStore/RestKit"
)

type PingResponse struct {
	Message string ` + "`" + `json:"message"` + "`" + `
}

func Ping() *restkit.EndpointRes[PingResponse] {
	return restkit.NewEndpointRes[PingResponse]().
		WithMethod(http.MethodGet).
		WithPath("/ping").
		WithTitle("Ping").
		WithDescription("Health check endpoint").
		WithHandler(func(ctx context.Context) (PingResponse, error) {
			return PingResponse{Message: "pong"}, nil
		})
}
`
