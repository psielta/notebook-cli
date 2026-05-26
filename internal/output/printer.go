package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"notebook-cli/internal/domain"
)

type Printer struct {
	w io.Writer
}

func NewPrinter(w io.Writer) *Printer {
	return &Printer{w: w}
}

func (p *Printer) PrintNotebooks(items []domain.NotebookListItem) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(p.w, "nenhum notebook encontrado")
		return err
	}

	tw := tabwriter.NewWriter(p.w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NOME\tNOTAS"); err != nil {
		return err
	}
	for _, item := range items {
		if _, err := fmt.Fprintf(tw, "%s\t%d\n", item.Name, item.NoteCount); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func (p *Printer) PrintNotes(notes []domain.Note) error {
	if len(notes) == 0 {
		_, err := fmt.Fprintln(p.w, "nenhuma nota")
		return err
	}

	tw := tabwriter.NewWriter(p.w, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "ID\tCRIADO EM\tTEXTO"); err != nil {
		return err
	}
	for _, note := range notes {
		createdAt := note.CreatedAt.Local().Format("2006-01-02 15:04:05")
		if _, err := fmt.Fprintf(tw, "%d\t%s\t%s\n", note.LocalID, createdAt, cleanText(note.Text)); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func cleanText(text string) string {
	replacer := strings.NewReplacer("\r", " ", "\n", " ", "\t", " ")
	return replacer.Replace(text)
}
