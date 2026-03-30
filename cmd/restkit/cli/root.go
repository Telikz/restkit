package cli

import (
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
)

type RootCmd struct{}

var (
	Title = lipgloss.NewStyle().
		Bold(true).Foreground(lipgloss.Color("4")).
		Render("RestKit CLI - Generate REST APIs")
	RootLongDesc = Title + `

This CLI helps you generate endpoint files for the restkit framework.`
)

func (r *RootCmd) Execute() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restkit",
		Short: Title,
		Long:  RootLongDesc,
	}

	cmd.CompletionOptions.HiddenDefaultCmd = true
	cmd.AddCommand((&AddCmd{}).Execute())
	cmd.AddCommand((&InitCmd{}).Execute())
	return cmd
}
