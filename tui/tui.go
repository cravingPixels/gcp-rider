// Package tui contains all the logic for the terminal user interface.
package tui

import (
	"context"
	"fmt"
	"gcp-rider/gcp"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// gcpClient is an interface that defines the methods we need from the gcp package.
// This allows us to use a mock client in our TUI tests.
type gcpClient interface {
	FetchInstances(ctx context.Context, projectID string) ([]gcp.Instance, error)
}

// Model represents the state of the TUI application.
type Model struct {
	gcpClient gcpClient
	projectID string
	vms       []gcp.Instance
	cursor    int
	loading   bool
	spinner   spinner.Model
	err       error
}

// vmsMsg is a message sent when the list of VMs has been fetched.
type vmsMsg []gcp.Instance

// errMsg is a message sent when an error occurs.
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

// NewModel creates a new TUI model with its dependencies.
func NewModel(client gcpClient, projectID string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return Model{
		gcpClient: client,
		projectID: projectID,
		loading:   true,
		spinner:   s,
	}
}

// Init is the first command run when the application starts.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetchVmsCmd)
}

// fetchVmsCmd is a command that fetches the VMs from GCP.
func (m Model) fetchVmsCmd() tea.Msg {
	vms, err := m.gcpClient.FetchInstances(context.Background(), m.projectID)
	if err != nil {
		return errMsg{err}
	}
	return vmsMsg(vms)
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.vms)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.vms) == 0 {
				return m, nil
			}
			vm := m.vms[m.cursor]
			cmd := exec.Command("gcloud", "compute", "ssh", vm.Name, "--zone", vm.Zone, "--project", m.projectID)
			return m, tea.ExecProcess(cmd, nil)
		}
	case vmsMsg:
		m.vms = msg
		m.loading = false
	case errMsg:
		m.err = msg
		m.loading = false
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

// View renders the user interface.
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nAn error occurred: %v\n\nPress q to quit.\n", m.err)
	}

	if m.loading {
		return fmt.Sprintf("\n %s Loading VMs...\n\n", m.spinner.View())
	}

	var b strings.Builder
	b.WriteString("GCP VMs:\n\n")
	for i, vm := range m.vms {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		b.WriteString(fmt.Sprintf("%s [%s]\n", cursor, vm.Name))
	}

	b.WriteString("\nPress q to quit.\n")
	return b.String()
}
