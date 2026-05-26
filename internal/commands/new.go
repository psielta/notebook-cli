package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"notebook-cli/internal/app"
)

func NewNewCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:  "new <nome>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			notebook, err := a.NotebookService.Create(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Notebook '%s' criado.\n", notebook.Name)
			return err
		},
	}
}
