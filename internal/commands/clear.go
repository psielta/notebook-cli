package commands

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"notebook-cli/internal/app"
)

func NewClearCmd(a *app.App) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:  "clear",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes {
				notes, err := a.NoteService.List(cmd.Context())
				if err != nil {
					return err
				}
				if len(notes) == 0 {
					_, err := fmt.Fprintln(cmd.OutOrStdout(), "0 notas removidas.")
					return err
				}

				if _, err := fmt.Fprint(cmd.ErrOrStderr(), "Remover todas as notas do notebook atual? [y/N] "); err != nil {
					return err
				}
				answer, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
				if err != nil && strings.TrimSpace(answer) == "" {
					return err
				}
				if strings.ToLower(strings.TrimSpace(answer)) != "y" {
					_, err := fmt.Fprintln(cmd.ErrOrStderr(), "Operacao cancelada.")
					return err
				}
			}

			count, err := a.NoteService.Clear(cmd.Context())
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "%d notas removidas.\n", count)
			return err
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "confirma a remocao sem prompt")
	return cmd
}
