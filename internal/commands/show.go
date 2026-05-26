package commands

import (
	"github.com/spf13/cobra"

	"notebook-cli/internal/app"
	"notebook-cli/internal/output"
)

func NewShowCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:  "show",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			notes, err := a.NoteService.List(cmd.Context())
			if err != nil {
				return err
			}
			return output.NewPrinter(cmd.OutOrStdout()).PrintNotes(notes)
		},
	}
}
