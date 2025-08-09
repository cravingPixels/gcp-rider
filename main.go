// Package main provides a simple terminal user interface (TUI) for viewing
// Google Cloud Platform (GCP) virtual machine instances.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"google.golang.org/api/iterator"
)

// vmInfo holds the information we need for each VM.
type vmInfo struct {
	name string
	zone string
}

// model represents the state of the application.
type model struct {
	vms       []vmInfo
	projectID string
	cursor    int
	loading   bool
	spinner   spinner.Model
	err       error
}

// vmsMsg is a message containing the list of VMs.
type vmsMsg []vmInfo

// errMsg is a message containing an error.
type errMsg struct{ err error }

// Error returns the error message.
func (e errMsg) Error() string { return e.err.Error() }

// fetchInstances retrieves a list of VMs from GCP for a given project.
func fetchInstances(ctx context.Context, client *compute.InstancesClient, projectID string) ([]vmInfo, error) {
	req := &computepb.AggregatedListInstancesRequest{
		Project: projectID,
	}
	it := client.AggregatedList(ctx, req)
	var vms []vmInfo
	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over instances: %w", err)
		}
		if pair.Value != nil && len(pair.Value.Instances) > 0 {
			for _, instance := range pair.Value.Instances {
				// The zone is a full URL, so we extract the last part.
				zone := path.Base(*instance.Zone)
				vms = append(vms, vmInfo{name: *instance.Name, zone: zone})
			}
		}
	}
	return vms, nil
}

// getGcpVmsCmd creates a Bubble Tea command that fetches the list of VMs.
func getGcpVmsCmd(projectID string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		client, err := compute.NewInstancesRESTClient(ctx)
		if err != nil {
			return errMsg{err}
		}
		defer client.Close()

		vms, err := fetchInstances(ctx, client, projectID)
		if err != nil {
			return errMsg{err}
		}
		return vmsMsg(vms)
	}
}

// initialModel returns the initial state of the application.
func initialModel(projectID string) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	return model{
		loading:   true,
		spinner:   s,
		projectID: projectID,
	}
}

// Init is the first command that is run when the application starts.
func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, getGcpVmsCmd(m.projectID))
}

// Update handles messages and updates the model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			cmd := exec.Command("gcloud", "compute", "ssh", vm.name, "--zone", vm.zone, "--project", m.projectID)
			return m, tea.ExecProcess(cmd, nil)
		}
	case vmsMsg:
		m.vms = msg
		m.loading = false
	case errMsg:
		m.err = msg
		m.loading = false
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

// View renders the user interface.
func (m model) View() string {
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
		b.WriteString(fmt.Sprintf("%s [%s]\n", cursor, vm.name))
	}

	b.WriteString("\nPress q to quit.\n")
	return b.String()
}

func main() {
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		fmt.Println("Error: GCP_PROJECT_ID environment variable not set.")
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel(projectID))
	if _, err := p.Run(); err != nil {
		log.Fatalf("Alas, there's been an error: %v", err)
	}
}

