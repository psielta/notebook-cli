package output

import (
	"io"
	"os"
)

type Style struct {
	color bool
}

func NewStyle(w io.Writer) Style {
	return NewStyleForTerminal(isTerminal(w))
}

func NewStyleForTerminal(terminal bool) Style {
	_, noColor := os.LookupEnv("NO_COLOR")
	return newStyle(terminal, noColor, os.Getenv("TERM"))
}

func newStyle(terminal, noColor bool, term string) Style {
	return Style{color: terminal && !noColor && term != "dumb"}
}

func (s Style) Enabled() bool {
	return s.color
}

func (s Style) Header(text string) string {
	return s.apply("1", text)
}

func (s Style) Success(text string) string {
	return s.apply("32", text)
}

func (s Style) Name(text string) string {
	return s.apply("36", text)
}

func (s Style) ID(text string) string {
	return s.apply("33", text)
}

func (s Style) Timestamp(text string) string {
	return s.apply("90", text)
}

func (s Style) apply(code, text string) string {
	if !s.color {
		return text
	}
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}

func isTerminal(w io.Writer) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}

	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
