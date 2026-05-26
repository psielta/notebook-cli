package commands

import (
	"github.com/spf13/cobra"

	"notebook-cli/internal/app"
	"notebook-cli/internal/version"
)

func NewRootCmd(a *app.App) *cobra.Command {
	root := &cobra.Command{
		Use:           "nb",
		Short:         "Gerencia notebooks de notas locais",
		Version:       version.Version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.SetVersionTemplate("{{.Name}} {{.Version}}\n")

	root.AddCommand(
		NewNewCmd(a),
		NewUseCmd(a),
		NewCurrentCmd(a),
		NewListCmd(a),
		NewAddCmd(a),
		NewShowCmd(a),
		NewLastCmd(a),
		NewRemoveCmd(a),
		NewClearCmd(a),
	)
	return root
}
