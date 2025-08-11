package tui

import (
	"errors"
	"gcp-rider/gcp"
	"gcp-rider/gcp/mocks"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdate_VMFetchSuccess(t *testing.T) {
	mockClient := new(mocks.Client)
	expectedVMs := []gcp.Instance{{Name: "vm-1", Zone: "z-1"}}
	mockClient.On("FetchInstances", mock.Anything, "test-project").Return(expectedVMs, nil)

	m := NewModel(mockClient, "test-project")

	msg := m.fetchVmsCmd()
	model, _ := m.Update(msg)
	updatedModel := model.(Model)

	require.False(t, updatedModel.loading, "expected loading to be false")
	require.Len(t, updatedModel.vms, 1, "expected 1 VM")
	require.Equal(t, "vm-1", updatedModel.vms[0].Name, "unexpected VM name")

	mockClient.AssertExpectations(t)
}

func TestUpdate_VMFetchError(t *testing.T) {
	mockClient := new(mocks.Client)
	expectedErr := errors.New("fetch failed")
	mockClient.On("FetchInstances", mock.Anything, "test-project").Return(nil, expectedErr)

	m := NewModel(mockClient, "test-project")

	msg := m.fetchVmsCmd()
	model, _ := m.Update(msg)
	updatedModel := model.(Model)

	require.False(t, updatedModel.loading, "expected loading to be false")
	require.Error(t, updatedModel.err, "expected an error")

	// Check that the underlying error matches our expected error.
	var e errMsg
	require.ErrorAs(t, updatedModel.err, &e, "error should be of type errMsg")
	require.Equal(t, expectedErr.Error(), e.err.Error(), "unexpected error message")

	mockClient.AssertExpectations(t)
}

func TestUpdate_CursorMovement(t *testing.T) {
	mockClient := new(mocks.Client)
	m := NewModel(mockClient, "")
	m.vms = []gcp.Instance{{Name: "vm-1"}, {Name: "vm-2"}, {Name: "vm-3"}}
	m.loading = false

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("down")})
	m = model.(Model)
	require.Equal(t, 1, m.cursor, "cursor should be 1 after moving down")

	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("up")})
	m = model.(Model)
	require.Equal(t, 0, m.cursor, "cursor should be 0 after moving up")
}