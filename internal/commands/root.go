package commands

import (
	"github.com/spf13/cobra"

	"notebook-cli/internal/app"
)

func NewRootCmd(a *app.App) *cobra.Command {
	root := &cobra.Command{
		Use:           "nb",
		Short:         "Gerencia notebooks de notas locais",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

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
