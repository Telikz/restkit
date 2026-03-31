package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type TemplateData struct {
	Package    string
	Name       string
	PascalName string
	LowerName  string
	Method     string
	Path       string
	Type       string
	Year       int
}

func GenerateEndpoint(
	name, endpointType, method, path, pkg, output string,
) error {
	if method == "" {
		switch endpointType {
		case "full", "req":
			method = "Post"
		case "res":
			method = "Get"
		default:
			method = "Post"
		}
	}

	if path == "" {
		path = "/" + strings.ToLower(name)
	}

	if err := os.MkdirAll(output, 0o755); err != nil {
		return err
	}

	data := TemplateData{
		Package:    pkg,
		Name:       name,
		PascalName: toPascalCase(name),
		LowerName:  strings.ToLower(name),
		Method:     toPascalCase(method),
		Path:       path,
		Type:       endpointType,
		Year:       time.Now().Year(),
	}

	var tmpl string
	switch endpointType {
	case "full":
		tmpl = fullEndpointTemplate
	case "req":
		tmpl = reqEndpointTemplate
	case "res":
		tmpl = resEndpointTemplate
	}

	filename := filepath.Join(output, strings.ToLower(name)+".go")

	if err := generateFile(filename, tmpl, data); err != nil {
		return err
	}

	exec.Command("gofmt", "-w", filename).Run()

	return nil
}

const fullEndpointTemplate = `package {{.Package}}

import (
	"context"
	"net/http"

	api "github.com/reststore/restkit"
)

type {{.PascalName}}Request struct {
	// Add your request fields here with validation tags from go-playground/validator
	// Example:
	// Name  string ` + "`" + `json:"name" validate:"required"` + "`" + `
	// Email string ` + "`" + `json:"email" validate:"required,email"` + "`" + `
}

type {{.PascalName}}Response struct {
	// Add your response fields here
}

// {{.PascalName}} creates a new full endpoint (request + response)
func {{.PascalName}}() *api.Endpoint[{{.PascalName}}Request, {{.PascalName}}Response] {
	return api.NewEndpoint[{{.PascalName}}Request, {{.PascalName}}Response]().
		WithMethod(http.Method{{.Method}}).
		WithPath("{{.Path}}").
		WithTitle("{{.PascalName}} Endpoint").
		WithDescription("Description for {{.LowerName}} endpoint").
		WithHandler(handle{{.PascalName}})
}

func handle{{.PascalName}}(ctx context.Context, req {{.PascalName}}Request) ({{.PascalName}}Response, error) {
	// TODO: Implement your handler logic here
	return {{.PascalName}}Response{}, nil
}
`

const reqEndpointTemplate = `package {{.Package}}

import (
	"context"
	"net/http"

	api "github.com/reststore/restkit"
)

type {{.PascalName}}Request struct {
	// Add your request fields here with validation tags from go-playground/validator
	// Example:
	// Name  string ` + "`" + `json:"name" validate:"required"` + "`" + `
	// Email string ` + "`" + `json:"email" validate:"required,email"` + "`" + `
}

// {{.PascalName}} creates a new request-only endpoint (returns 204 No Content)
func {{.PascalName}}() *api.EndpointReq[{{.PascalName}}Request] {
	return api.NewEndpointReq[{{.PascalName}}Request]().
		WithMethod(http.Method{{.Method}}).
		WithPath("{{.Path}}").
		WithTitle("{{.PascalName}} Endpoint").
		WithDescription("Description for {{.LowerName}} endpoint").
		WithHandler(handle{{.PascalName}})
}

func handle{{.PascalName}}(ctx context.Context, req {{.PascalName}}Request) error {
	// TODO: Implement your handler logic here
	return nil
}
`

const resEndpointTemplate = `package {{.Package}}

import (
	"context"
	"net/http"

	api "github.com/reststore/restkit"
)

type {{.PascalName}}Response struct {
	// Add your response fields here
}

// {{.PascalName}} creates a new response-only endpoint (no request body)
func {{.PascalName}}() *api.EndpointRes[{{.PascalName}}Response] {
	return api.NewEndpointRes[{{.PascalName}}Response]().
		WithMethod(http.Method{{.Method}}).
		WithPath("{{.Path}}").
		WithTitle("{{.PascalName}} Endpoint").
		WithDescription("Description for {{.LowerName}} endpoint").
		WithHandler(handle{{.PascalName}})
}

func handle{{.PascalName}}(ctx context.Context) ({{.PascalName}}Response, error) {
	// TODO: Implement your handler logic here
	return {{.PascalName}}Response{}, nil
}
`
