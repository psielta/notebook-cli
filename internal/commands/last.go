package commands

import (
	"strconv"

	"github.com/spf13/cobra"

	"notebook-cli/internal/app"
	"notebook-cli/internal/apperrors"
	"notebook-cli/internal/output"
)

func NewLastCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:  "last [n]",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n := 1
			if len(args) == 1 {
				parsed, err := strconv.Atoi(args[0])
				if err != nil || parsed <= 0 {
					return apperrors.ErrInvalidCount
				}
				n = parsed
			}

			notes, err := a.NoteService.Last(cmd.Context(), n)
			if err != nil {
				return err
			}
			return output.NewPrinter(cmd.OutOrStdout()).PrintNotes(notes)
		},
	}
}
