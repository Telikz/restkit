package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/RestStore/RestKit/cmd/restkit/generator"
)

type AddCmd struct {
	endpointType string
	method       string
	path         string
	pkg          string
	output       string
}

const (
	AddShortDesc = "Add a new endpoint"
	AddLongDesc  = `Add a new endpoint with the specified name.

The endpoint name will be used to generate appropriate Go types and handlers.
Use flags to customize the endpoint type, HTTP method, URL path, package name, and output directory.`

	AddExample = `restkit add user -t full -p /users
restkit add health -t res -m GET
restkit add delete -t req -m DELETE`
)

func (a *AddCmd) Execute() *cobra.Command {
	addCmd := &cobra.Command{
		Use: "add [name]",

		Short: AddShortDesc,
		Long:  AddLongDesc,

		Example: AddExample,

		Args: cobra.ExactArgs(1),
		RunE: a.handleGenEndpoint,
	}

	addCmd.Flags().
		StringVarP(&a.endpointType, "type", "t", "full", "Endpoint type")
	addCmd.Flags().StringVarP(&a.method, "method", "m", "", "HTTP method")
	addCmd.Flags().StringVarP(&a.path, "path", "p", "", "URL path")
	addCmd.Flags().
		StringVarP(&a.pkg, "package", "k", "endpoints", "Package name")
	addCmd.Flags().
		StringVarP(&a.output, "output", "o", "./endpoints", "Output directory")

	return addCmd
}

func (a *AddCmd) handleGenEndpoint(cmd *cobra.Command, args []string) error {
	name := args[0]
	if name == "" {
		return fmt.Errorf("endpoint name is required")
	}

	if !validEndpointType(a.endpointType) {
		return fmt.Errorf(
			"invalid endpoint type %q, must be one of: full, req, res",
			a.endpointType,
		)
	}

	if err := generator.GenerateEndpoint(
		name, a.endpointType, a.method, a.path, a.pkg, a.output,
	); err != nil {
		return err
	}

	fmt.Printf("Generated endpoint: %s/%s.go (%s)\n",
		a.output, name, a.endpointType)
	return nil
}

func validEndpointType(endpointType string) bool {
	return endpointType == "full" || endpointType == "req" ||
		endpointType == "res"
}
