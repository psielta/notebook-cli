package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"notebook-cli/internal/apperrors"
	"notebook-cli/internal/domain"
)

func TestNotebookServiceCreate(t *testing.T) {
	t.Run("creates notebook", func(t *testing.T) {
		repo := newFakeNotebookRepo()
		svc := NewNotebookService(repo, newFakeCurrentStore())

		notebook, err := svc.Create(context.Background(), "erp")

		require.NoError(t, err)
		require.Equal(t, "erp", notebook.Name)
		require.Equal(t, 1, notebook.NextNoteID)
	})

	t.Run("rejects invalid name", func(t *testing.T) {
		svc := NewNotebookService(newFakeNotebookRepo(), newFakeCurrentStore())

		_, err := svc.Create(context.Background(), "meu erp")

		require.ErrorIs(t, err, apperrors.ErrInvalidName)
	})

	t.Run("rejects duplicate", func(t *testing.T) {
		repo := newFakeNotebookRepo()
		repo.notebooks["erp"] = &domain.Notebook{ID: 1, Name: "erp", NextNoteID: 1}
		svc := NewNotebookService(repo, newFakeCurrentStore())

		_, err := svc.Create(context.Background(), "erp")

		require.ErrorIs(t, err, apperrors.ErrNotebookExists)
	})
}

func TestNotebookServiceCreateTask(t *testing.T) {
	repo := newFakeNotebookRepo()
	current := newFakeCurrentStore()
	svc := newNotebookService(repo, current, fixedTaskNow)

	notebook, err := svc.CreateTask(context.Background())

	require.NoError(t, err)
	require.Equal(t, "T001-2026-05-27", notebook.Name)
	require.Equal(t, 1, notebook.NextNoteID)
	require.Equal(t, "T001-2026-05-27", current.name)
}

func TestNotebookServiceCreateTaskUsesNextDailyNumber(t *testing.T) {
	repo := newFakeNotebookRepo()
	repo.notebooks["T001-2026-05-27"] = &domain.Notebook{ID: 1, Name: "T001-2026-05-27", NextNoteID: 1}
	repo.notebooks["T003-2026-05-27"] = &domain.Notebook{ID: 2, Name: "T003-2026-05-27", NextNoteID: 1}
	repo.notebooks["T099-2026-05-26"] = &domain.Notebook{ID: 3, Name: "T099-2026-05-26", NextNoteID: 1}
	repo.notebooks["erp"] = &domain.Notebook{ID: 4, Name: "erp", NextNoteID: 1}
	svc := newNotebookService(repo, newFakeCurrentStore(), fixedTaskNow)

	notebook, err := svc.CreateTask(context.Background())

	require.NoError(t, err)
	require.Equal(t, "T004-2026-05-27", notebook.Name)
}

func TestNotebookServiceCreateTaskLimit(t *testing.T) {
	repo := newFakeNotebookRepo()
	repo.notebooks["T999-2026-05-27"] = &domain.Notebook{ID: 1, Name: "T999-2026-05-27", NextNoteID: 1}
	svc := newNotebookService(repo, newFakeCurrentStore(), fixedTaskNow)

	_, err := svc.CreateTask(context.Background())

	require.ErrorIs(t, err, apperrors.ErrTaskLimitReached)
}

func TestIsTaskTemplateName(t *testing.T) {
	require.True(t, IsTaskTemplateName("task"))
	require.True(t, IsTaskTemplateName(" task "))
	require.False(t, IsTaskTemplateName("Task"))
	require.False(t, IsTaskTemplateName("tarefa"))
}

func TestNotebookServiceUseAndCurrent(t *testing.T) {
	repo := newFakeNotebookRepo()
	repo.notebooks["erp"] = &domain.Notebook{ID: 1, Name: "erp", NextNoteID: 1}
	current := newFakeCurrentStore()
	svc := NewNotebookService(repo, current)

	_, err := svc.Use(context.Background(), "erp")
	require.NoError(t, err)
	require.Equal(t, "erp", current.name)

	name, err := svc.Current(context.Background())
	require.NoError(t, err)
	require.Equal(t, "erp", name)
}

func TestNotebookServiceUseMissing(t *testing.T) {
	svc := NewNotebookService(newFakeNotebookRepo(), newFakeCurrentStore())

	_, err := svc.Use(context.Background(), "missing")

	require.ErrorIs(t, err, apperrors.ErrNotebookNotFound)
}

func TestNotebookServiceUseInvalidNameAndSetError(t *testing.T) {
	repo := newFakeNotebookRepo()
	repo.notebooks["erp"] = &domain.Notebook{ID: 1, Name: "erp", NextNoteID: 1}

	svc := NewNotebookService(repo, newFakeCurrentStore())
	_, err := svc.Use(context.Background(), "nome invalido")
	require.ErrorIs(t, err, apperrors.ErrInvalidName)

	svc = NewNotebookService(repo, &fakeCurrentStore{err: errFake})
	_, err = svc.Use(context.Background(), "erp")
	require.ErrorIs(t, err, errFake)
}

func TestNotebookServiceListIncludesNoteCounts(t *testing.T) {
	repo := newFakeNotebookRepo()
	repo.notebooks["erp"] = &domain.Notebook{
		ID:         1,
		Name:       "erp",
		NextNoteID: 3,
		Notes:      []domain.Note{{LocalID: 1}, {LocalID: 2}},
	}
	svc := NewNotebookService(repo, newFakeCurrentStore())

	items, err := svc.List(context.Background())

	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "erp", items[0].Name)
	require.Equal(t, 2, items[0].NoteCount)
}

func TestNotebookServiceCurrentEmpty(t *testing.T) {
	svc := NewNotebookService(newFakeNotebookRepo(), newFakeCurrentStore())

	_, err := svc.Current(context.Background())

	require.ErrorIs(t, err, apperrors.ErrNoCurrentNotebook)
}

func TestNotebookServiceCurrentStaleNotebook(t *testing.T) {
	current := &fakeCurrentStore{name: "stale"}
	svc := NewNotebookService(newFakeNotebookRepo(), current)

	_, err := svc.Current(context.Background())

	require.ErrorIs(t, err, apperrors.ErrNotebookNotFound)
}

func TestNotebookServicePropagatesCurrentStoreError(t *testing.T) {
	current := &fakeCurrentStore{err: errFake}
	svc := NewNotebookService(newFakeNotebookRepo(), current)

	_, err := svc.Current(context.Background())

	require.ErrorIs(t, err, errFake)
}

type fakeNotebookRepo struct {
	nextID    uint
	notebooks map[string]*domain.Notebook
	err       error
}

func newFakeNotebookRepo() *fakeNotebookRepo {
	return &fakeNotebookRepo{nextID: 1, notebooks: make(map[string]*domain.Notebook)}
}

func (f *fakeNotebookRepo) Create(ctx context.Context, notebook *domain.Notebook) error {
	if f.err != nil {
		return f.err
	}
	if _, exists := f.notebooks[notebook.Name]; exists {
		return apperrors.ErrNotebookExists
	}
	notebook.ID = f.nextID
	f.nextID++
	f.notebooks[notebook.Name] = notebook
	return nil
}

func (f *fakeNotebookRepo) GetByName(ctx context.Context, name string) (*domain.Notebook, error) {
	if f.err != nil {
		return nil, f.err
	}
	notebook, ok := f.notebooks[name]
	if !ok {
		return nil, apperrors.ErrNotebookNotFound
	}
	return notebook, nil
}

func (f *fakeNotebookRepo) List(ctx context.Context) ([]domain.NotebookListItem, error) {
	if f.err != nil {
		return nil, f.err
	}
	items := make([]domain.NotebookListItem, 0, len(f.notebooks))
	for _, notebook := range f.notebooks {
		items = append(items, domain.NotebookListItem{
			Name:      notebook.Name,
			NoteCount: len(notebook.Notes),
		})
	}
	return items, nil
}

type fakeCurrentStore struct {
	name string
	err  error
}

func newFakeCurrentStore() *fakeCurrentStore {
	return &fakeCurrentStore{}
}

func (f *fakeCurrentStore) Get() (string, error) {
	return f.name, f.err
}

func (f *fakeCurrentStore) Set(name string) error {
	if f.err != nil {
		return f.err
	}
	f.name = name
	return nil
}

func (f *fakeCurrentStore) Clear() error {
	if f.err != nil {
		return f.err
	}
	f.name = ""
	return nil
}

var errFake = errors.New("fake error")

func fixedTaskNow() time.Time {
	return time.Date(2026, 5, 27, 12, 0, 0, 0, time.Local)
}
