package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/telikz/restkit/cmd/restkit/generator"
)

type InitCmd struct{}

const (
	InitShortDesc = "Initialize a new project"
	InitLongDesc  = `Initialize a new project

This command sets up a new Go module with the necessary structure for a restkit API.`
)

func (i *InitCmd) Execute() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init [module-name]",
		Short: InitShortDesc,
		Long:  InitLongDesc,
		Args:  cobra.MaximumNArgs(1),
		RunE:  i.handleInit,
	}

	return initCmd
}

func (i *InitCmd) handleInit(cmd *cobra.Command, args []string) error {
	moduleName := ""
	if len(args) > 0 {
		moduleName = args[0]
	}

	if err := generator.InitProject(moduleName); err != nil {
		return err
	}

	fmt.Println("Project initialized successfully!")
	return nil
}
