package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"notebook-cli/internal/app"
)

func NewAddCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:  "add <texto>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			note, err := a.NoteService.Add(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Nota %d adicionada.\n", note.LocalID)
			return err
		},
	}
}
