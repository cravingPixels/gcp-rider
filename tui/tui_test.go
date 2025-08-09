package tui

import (
	"context"
	"errors"
	"gcp-rider/gcp"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// mockGcpClient is a mock implementation of the gcpClient interface for TUI tests.
type mockGcpClient struct {
	instances []gcp.Instance
	err       error
}

func (m *mockGcpClient) FetchInstances(ctx context.Context, projectID string) ([]gcp.Instance, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.instances, nil
}

func TestUpdate_VMFetchSuccess(t *testing.T) {
	client := &mockGcpClient{
		instances: []gcp.Instance{{Name: "vm-1", Zone: "z-1"}},
	}
	m := NewModel(client, "test-project")

	// Simulate the initial fetch command
	msg := m.fetchVmsCmd()
	model, _ := m.Update(msg)
	updatedModel := model.(Model)

	if updatedModel.loading {
		t.Error("expected loading to be false after fetching VMs")
	}
	if len(updatedModel.vms) != 1 {
		t.Errorf("expected 1 VM, got %d", len(updatedModel.vms))
	}
	if updatedModel.vms[0].Name != "vm-1" {
		t.Errorf("unexpected VM name: %s", updatedModel.vms[0].Name)
	}
}

func TestUpdate_VMFetchError(t *testing.T) {
	expectedErr := errors.New("fetch failed")
	client := &mockGcpClient{err: expectedErr}
	m := NewModel(client, "test-project")

	msg := m.fetchVmsCmd()
	model, _ := m.Update(msg)
	updatedModel := model.(Model)

	if updatedModel.loading {
		t.Error("expected loading to be false after an error")
	}
	if updatedModel.err == nil {
		t.Error("expected an error but got nil")
	}
	if !errors.Is(updatedModel.err, expectedErr) {
		t.Errorf("expected error '%v', got '%v'", expectedErr, updatedModel.err)
	}
}

func TestUpdate_CursorMovement(t *testing.T) {
	m := Model{
		vms: []gcp.Instance{
			{Name: "vm-1"},
			{Name: "vm-2"},
			{Name: "vm-3"},
		},
		cursor: 0,
	}

	// Move down
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("down")})
	m = model.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor to be 1 after moving down, got %d", m.cursor)
	}

	// Move up
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("up")})
	m = model.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor to be 0 after moving up, got %d", m.cursor)
	}

	// Test boundary
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("up")})
	m = model.(Model)
	if m.cursor != 0 {
		t.Errorf("expected cursor to stay at 0 at the top boundary, got %d", m.cursor)
	}
}
