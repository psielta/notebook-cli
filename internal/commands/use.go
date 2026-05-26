package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"notebook-cli/internal/app"
)

func NewUseCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:  "use <nome>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			notebook, err := a.NotebookService.Use(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Usando '%s'.\n", notebook.Name)
			return err
		},
	}
}
