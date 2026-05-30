package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/riccardo/dotty/internal/app"
	"github.com/riccardo/dotty/internal/scan"
)

func main() {
	homeFlag := flag.String("home", "", "Override home directory for scanning (useful for testing)")
	flag.Parse()

	homeDir := *homeFlag
	if homeDir == "" {
		var err error
		homeDir, err = os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to determine home directory: %v\n", err)
			os.Exit(1)
		}
	}

	roots, total, err := scan.Dotfiles(homeDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to scan dotfiles: %v\n", err)
		os.Exit(1)
	}
	if total == 0 {
		fmt.Println("No dotfiles found in ~")
		os.Exit(1)
	}

	m := app.New(homeDir, roots, total)
	p := tea.NewProgram(
		&m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running dotty: %v\n", err)
		os.Exit(1)
	}
}
