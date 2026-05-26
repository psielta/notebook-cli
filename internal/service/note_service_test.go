package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"notebook-cli/internal/apperrors"
	"notebook-cli/internal/domain"
)

func TestNoteServiceAdd(t *testing.T) {
	notebooks := newFakeNotebookRepo()
	notebooks.notebooks["erp"] = &domain.Notebook{ID: 1, Name: "erp", NextNoteID: 1}
	current := &fakeCurrentStore{name: "erp"}
	notes := newFakeNoteRepo()
	svc := NewNoteService(notes, notebooks, current)

	first, err := svc.Add(context.Background(), "corrigir problema x")
	require.NoError(t, err)
	second, err := svc.Add(context.Background(), "testar importacao")
	require.NoError(t, err)

	require.Equal(t, 1, first.LocalID)
	require.Equal(t, 2, second.LocalID)
}

func TestNoteServiceAddErrors(t *testing.T) {
	t.Run("empty text", func(t *testing.T) {
		svc := NewNoteService(newFakeNoteRepo(), newFakeNotebookRepo(), newFakeCurrentStore())

		_, err := svc.Add(context.Background(), " ")

		require.ErrorIs(t, err, apperrors.ErrEmptyText)
	})

	t.Run("no current", func(t *testing.T) {
		svc := NewNoteService(newFakeNoteRepo(), newFakeNotebookRepo(), newFakeCurrentStore())

		_, err := svc.Add(context.Background(), "x")

		require.ErrorIs(t, err, apperrors.ErrNoCurrentNotebook)
	})
}

func TestNoteServiceListLastRemoveClear(t *testing.T) {
	notebooks := newFakeNotebookRepo()
	notebooks.notebooks["erp"] = &domain.Notebook{ID: 1, Name: "erp", NextNoteID: 1}
	current := &fakeCurrentStore{name: "erp"}
	notes := newFakeNoteRepo()
	svc := NewNoteService(notes, notebooks, current)

	_, _ = svc.Add(context.Background(), "um")
	_, _ = svc.Add(context.Background(), "dois")
	_, _ = svc.Add(context.Background(), "tres")

	all, err := svc.List(context.Background())
	require.NoError(t, err)
	require.Len(t, all, 3)

	last, err := svc.Last(context.Background(), 2)
	require.NoError(t, err)
	require.Equal(t, []int{2, 3}, []int{last[0].LocalID, last[1].LocalID})

	require.NoError(t, svc.Remove(context.Background(), 2))
	remaining, err := svc.List(context.Background())
	require.NoError(t, err)
	require.Equal(t, []int{1, 3}, []int{remaining[0].LocalID, remaining[1].LocalID})

	count, err := svc.Clear(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(2), count)
}

func TestNoteServiceRemoveMissingAndInvalidLast(t *testing.T) {
	notebooks := newFakeNotebookRepo()
	notebooks.notebooks["erp"] = &domain.Notebook{ID: 1, Name: "erp", NextNoteID: 1}
	current := &fakeCurrentStore{name: "erp"}
	svc := NewNoteService(newFakeNoteRepo(), notebooks, current)

	err := svc.Remove(context.Background(), 99)
	require.ErrorIs(t, err, apperrors.ErrNoteNotFound)

	_, err = svc.Last(context.Background(), 0)
	require.ErrorIs(t, err, apperrors.ErrInvalidCount)
}

func TestNoteServiceInvalidRemoveID(t *testing.T) {
	svc := NewNoteService(newFakeNoteRepo(), newFakeNotebookRepo(), newFakeCurrentStore())

	err := svc.Remove(context.Background(), 0)

	require.ErrorIs(t, err, apperrors.ErrInvalidID)
}

func TestNoteServiceCurrentNotebookMissing(t *testing.T) {
	notebooks := newFakeNotebookRepo()
	current := &fakeCurrentStore{name: "stale"}
	svc := NewNoteService(newFakeNoteRepo(), notebooks, current)

	_, err := svc.List(context.Background())

	require.ErrorIs(t, err, apperrors.ErrNotebookNotFound)
}

func TestNoteServiceClearNoCurrent(t *testing.T) {
	svc := NewNoteService(newFakeNoteRepo(), newFakeNotebookRepo(), newFakeCurrentStore())

	_, err := svc.Clear(context.Background())

	require.ErrorIs(t, err, apperrors.ErrNoCurrentNotebook)
}

func TestNoteServiceRepositoryErrors(t *testing.T) {
	notebooks := newFakeNotebookRepo()
	notebooks.notebooks["erp"] = &domain.Notebook{ID: 1, Name: "erp", NextNoteID: 1}
	current := &fakeCurrentStore{name: "erp"}
	notes := newFakeNoteRepo()
	notes.err = errFake
	svc := NewNoteService(notes, notebooks, current)

	_, err := svc.Add(context.Background(), "x")
	require.ErrorIs(t, err, errFake)

	_, err = svc.List(context.Background())
	require.ErrorIs(t, err, errFake)

	_, err = svc.Last(context.Background(), 1)
	require.ErrorIs(t, err, errFake)

	err = svc.Remove(context.Background(), 1)
	require.ErrorIs(t, err, errFake)

	_, err = svc.Clear(context.Background())
	require.ErrorIs(t, err, errFake)
}

type fakeNoteRepo struct {
	nextID uint
	notes  map[uint][]domain.Note
	err    error
}

func newFakeNoteRepo() *fakeNoteRepo {
	return &fakeNoteRepo{nextID: 1, notes: make(map[uint][]domain.Note)}
}

func (f *fakeNoteRepo) Create(ctx context.Context, notebookID uint, text string) (*domain.Note, error) {
	if f.err != nil {
		return nil, f.err
	}
	localID := len(f.notes[notebookID]) + 1
	note := domain.Note{
		ID:         f.nextID,
		NotebookID: notebookID,
		LocalID:    localID,
		Text:       text,
		CreatedAt:  time.Now().UTC(),
	}
	f.nextID++
	f.notes[notebookID] = append(f.notes[notebookID], note)
	return &note, nil
}

func (f *fakeNoteRepo) ListByNotebook(ctx context.Context, notebookID uint) ([]domain.Note, error) {
	if f.err != nil {
		return nil, f.err
	}
	return append([]domain.Note(nil), f.notes[notebookID]...), nil
}

func (f *fakeNoteRepo) ListLastByNotebook(ctx context.Context, notebookID uint, n int) ([]domain.Note, error) {
	if f.err != nil {
		return nil, f.err
	}
	notes := f.notes[notebookID]
	if n > len(notes) {
		n = len(notes)
	}
	return append([]domain.Note(nil), notes[len(notes)-n:]...), nil
}

func (f *fakeNoteRepo) DeleteByLocalID(ctx context.Context, notebookID uint, localID int) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	notes := f.notes[notebookID]
	for i, note := range notes {
		if note.LocalID == localID {
			f.notes[notebookID] = append(notes[:i], notes[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeNoteRepo) DeleteAllByNotebook(ctx context.Context, notebookID uint) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	count := int64(len(f.notes[notebookID]))
	f.notes[notebookID] = nil
	return count, nil
}
