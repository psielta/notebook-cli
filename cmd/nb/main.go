package main

import (
	"fmt"
	"io"
	"os"

	"notebook-cli/internal/app"
	"notebook-cli/internal/commands"
	"notebook-cli/internal/config"
	"notebook-cli/internal/database"
	"notebook-cli/internal/repository"
	"notebook-cli/internal/service"
	"notebook-cli/internal/session"
	"notebook-cli/internal/state"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(stderr, commands.FormatError(err))
		return 2
	}

	db, err := database.Open(cfg.DBPath)
	if err != nil {
		fmt.Fprintln(stderr, commands.FormatError(err))
		return 2
	}
	if sqlDB, err := db.DB(); err == nil {
		defer sqlDB.Close()
	}

	sessionID := session.DefaultResolver().Resolve()
	_ = state.MigrateLegacy(cfg.BaseDir, sessionID)
	_ = state.MaybePrune(cfg.BaseDir, session.IsAlive)

	notebookRepo := repository.NewNotebookRepository(db)
	noteRepo := repository.NewNoteRepository(db)
	current := state.New(cfg.BaseDir, sessionID)

	a := &app.App{
		DB:              db,
		NotebookService: service.NewNotebookService(notebookRepo, current),
		NoteService:     service.NewNoteService(noteRepo, notebookRepo, current),
		CurrentStore:    current,
	}

	root := commands.NewRootCmd(a)
	root.SetArgs(args)
	root.SetIn(stdin)
	root.SetOut(stdout)
	root.SetErr(stderr)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(root.ErrOrStderr(), commands.FormatError(err))
		if commands.IsBusinessError(err) {
			return 1
		}
		return 2
	}
	return 0
}
