package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"notebook-cli/internal/app"
	"notebook-cli/internal/apperrors"
)

func NewRemoveCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:  "remove <id>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			localID, err := strconv.Atoi(args[0])
			if err != nil {
				return apperrors.ErrInvalidID
			}
			if err := a.NoteService.Remove(cmd.Context(), localID); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Nota %d removida.\n", localID)
			return err
		},
	}
}
