package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"notebook-cli/internal/app"
	"notebook-cli/internal/output"
	"notebook-cli/internal/service"
)

func NewNewCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:  "new <nome>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			style := output.NewStyle(cmd.OutOrStdout())
			if service.IsTaskTemplateName(args[0]) {
				notebook, err := a.NotebookService.CreateTask(cmd.Context())
				if err != nil {
					return err
				}
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s '%s' criado e selecionado.\n", style.Success("Notebook"), style.Name(notebook.Name))
				return err
			}

			notebook, err := a.NotebookService.Create(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s '%s' criado.\n", style.Success("Notebook"), style.Name(notebook.Name))
			return err
		},
	}
}
