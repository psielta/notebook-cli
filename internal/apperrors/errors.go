package apperrors

import "errors"

var (
	ErrNotebookNotFound  = errors.New("notebook nao encontrado")
	ErrNotebookExists    = errors.New("notebook ja existe")
	ErrInvalidName       = errors.New("nome de notebook invalido (use [a-zA-Z0-9_-], 1-100 chars)")
	ErrNoCurrentNotebook = errors.New("nenhum notebook selecionado (use `nb use <nome>`)")
	ErrNoteNotFound      = errors.New("nota nao encontrada")
	ErrEmptyText         = errors.New("texto da nota nao pode ser vazio")
	ErrInvalidID         = errors.New("id invalido")
	ErrInvalidCount      = errors.New("quantidade invalida")
)
