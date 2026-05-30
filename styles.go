package main

import "github.com/charmbracelet/lipgloss"

// Foreground-only styles: terminal background shows through everywhere.
var (
	colorMuted  = lipgloss.Color("244")
	colorAccent = lipgloss.Color("178")
	colorMatch  = lipgloss.Color("142")
	colorError  = lipgloss.Color("196")
)

var (
	styleActiveBorder = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorAccent)

	styleInactiveBorder = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorMuted)

	styleActiveItem = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent)

	styleDirItem = lipgloss.NewStyle().
			Foreground(colorAccent)

	styleStatusCount = lipgloss.NewStyle().
				Foreground(colorMuted)

	styleStatusCountNum = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	styleStatusKey = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	styleStatusLabel = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleStatusSep = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleStatusMsg = lipgloss.NewStyle().
			Foreground(colorMatch).
			Bold(true)

	styleLineNumber = lipgloss.NewStyle().
			Foreground(colorMuted).
			Width(4).
			Align(lipgloss.Right)

	styleComment = lipgloss.NewStyle().
			Foreground(colorMuted)

	styleError = lipgloss.NewStyle().
			Foreground(colorError)

	styleDim = lipgloss.NewStyle().
			Foreground(colorMuted)

	stylePanelTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent)

	stylePickerBorder = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorMuted).
				Padding(0, 1)

	stylePickerSelected = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent)

	stylePickerMatch = lipgloss.NewStyle().
				Foreground(colorMatch)

	stylePickerFooter = lipgloss.NewStyle().
				Foreground(colorMuted).
				Italic(true)

	stylePickerDim = lipgloss.NewStyle().
			Foreground(colorMuted)

	stylePickerDivider = lipgloss.NewStyle().
				Foreground(colorMuted)
)
