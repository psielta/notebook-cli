package output

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"notebook-cli/internal/domain"
)

type Printer struct {
	w     io.Writer
	style Style
}

func NewPrinter(w io.Writer) *Printer {
	return &Printer{w: w, style: NewStyle(w)}
}

func (p *Printer) PrintNotebooks(items []domain.NotebookListItem) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(p.w, "nenhum notebook encontrado")
		return err
	}

	rows := [][]tableCell{
		{
			{raw: "NOME", styled: p.style.Header("NOME")},
			{raw: "NOTAS", styled: p.style.Header("NOTAS")},
		},
	}
	for _, item := range items {
		noteCount := strconv.Itoa(item.NoteCount)
		rows = append(rows, []tableCell{
			{raw: item.Name, styled: p.style.Name(item.Name)},
			{raw: noteCount, styled: noteCount},
		})
	}
	return printTable(p.w, rows)
}

func (p *Printer) PrintNotes(notes []domain.Note) error {
	if len(notes) == 0 {
		_, err := fmt.Fprintln(p.w, "nenhuma nota")
		return err
	}

	rows := [][]tableCell{
		{
			{raw: "ID", styled: p.style.Header("ID")},
			{raw: "CRIADO EM", styled: p.style.Header("CRIADO EM")},
			{raw: "TEXTO", styled: p.style.Header("TEXTO")},
		},
	}
	for _, note := range notes {
		localID := strconv.Itoa(note.LocalID)
		createdAt := note.CreatedAt.Local().Format("2006-01-02 15:04:05")
		rows = append(rows, []tableCell{
			{raw: localID, styled: p.style.ID(localID)},
			{raw: createdAt, styled: p.style.Timestamp(createdAt)},
			{raw: CleanText(note.Text), styled: CleanText(note.Text)},
		})
	}
	return printTable(p.w, rows)
}

func (p *Printer) PrintAddedNote(note domain.Note) error {
	createdAt := note.CreatedAt.Local().Format("2006-01-02 15:04:05")
	_, err := fmt.Fprintf(
		p.w,
		"#%s %s  %s\n",
		p.style.ID(strconv.Itoa(note.LocalID)),
		p.style.Timestamp(createdAt),
		CleanText(note.Text),
	)
	return err
}

func CleanText(text string) string {
	replacer := strings.NewReplacer("\r", " ", "\n", " ", "\t", " ")
	return replacer.Replace(text)
}

type tableCell struct {
	raw    string
	styled string
}

func printTable(w io.Writer, rows [][]tableCell) error {
	widths := tableWidths(rows)
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				if _, err := fmt.Fprint(w, "  "); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprint(w, paddedCell(cell, widths[i])); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}
	return nil
}

func tableWidths(rows [][]tableCell) []int {
	var widths []int
	for _, row := range rows {
		for i, cell := range row {
			for len(widths) <= i {
				widths = append(widths, 0)
			}
			if len(cell.raw) > widths[i] {
				widths[i] = len(cell.raw)
			}
		}
	}
	return widths
}

func paddedCell(cell tableCell, width int) string {
	padding := width - len(cell.raw)
	if padding <= 0 {
		return cell.styled
	}
	return cell.styled + strings.Repeat(" ", padding)
}
