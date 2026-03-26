package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/BlackMetalz/vault-vim/internal/config"
	"github.com/BlackMetalz/vault-vim/internal/ui"
	"github.com/BlackMetalz/vault-vim/internal/vault"
)

var Version = "0.0.1-dev"

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nSet VAULT_ADDR and VAULT_TOKEN, or run 'vault login' first.\n")
		os.Exit(1)
	}

	client := vault.NewClient(cfg)
	app := ui.NewApp(client, Version)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running vault-vim: %v\n", err)
		os.Exit(1)
	}
}
