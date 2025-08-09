// Package main is the entry point for the gcp-rider application.
package main

import (
	"context"
	"fmt"
	"gcp-rider/gcp"
	"gcp-rider/tui"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		fmt.Println("Error: GCP_PROJECT_ID environment variable not set.")
		os.Exit(1)
	}

	// Create the real GCP client.
	gcpClient, err := gcp.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create GCP client: %v", err)
	}

	// Create the TUI model, injecting the GCP client as a dependency.
	tuiModel := tui.NewModel(gcpClient, projectID)

	// Start the Bubble Tea program.
	p := tea.NewProgram(tuiModel)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Alas, there's been an error: %v", err)
	}
}