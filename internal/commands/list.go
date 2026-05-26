package commands

import (
	"github.com/spf13/cobra"

	"notebook-cli/internal/app"
	"notebook-cli/internal/output"
)

func NewListCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:  "list",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			items, err := a.NotebookService.List(cmd.Context())
			if err != nil {
				return err
			}
			return output.NewPrinter(cmd.OutOrStdout()).PrintNotebooks(items)
		},
	}
}
