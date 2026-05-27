package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"notebook-cli/internal/app"
	"notebook-cli/internal/output"
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
			style := output.NewStyle(cmd.OutOrStdout())
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s '%s'.\n", style.Success("Usando"), style.Name(notebook.Name))
			return err
		},
	}
}
