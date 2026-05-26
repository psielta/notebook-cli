package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"notebook-cli/internal/app"
)

func NewCurrentCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:  "current",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := a.NotebookService.Current(cmd.Context())
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), name)
			return err
		},
	}
}
