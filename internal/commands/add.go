package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"notebook-cli/internal/app"
	"notebook-cli/internal/output"
	"notebook-cli/internal/shortcuts"
)

func NewAddCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:  "add <texto>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := args[0]
			if resolved, ok := shortcuts.ResolveNote(text); ok {
				text = resolved
			}

			note, err := a.NoteService.Add(cmd.Context(), text)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			style := output.NewStyle(out)
			if _, err := fmt.Fprintf(out, "%s %s adicionada.\n", style.Success("Nota"), style.ID(strconv.Itoa(note.LocalID))); err != nil {
				return err
			}
			return output.NewPrinter(out).PrintAddedNote(*note)
		},
	}
}
