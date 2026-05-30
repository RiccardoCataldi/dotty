package ui

import "github.com/charmbracelet/lipgloss"

// Foreground-only styles: terminal background shows through everywhere.
var (
	ColorMuted  = lipgloss.Color("244")
	ColorAccent = lipgloss.Color("178")
	ColorMatch  = lipgloss.Color("142")
	ColorError  = lipgloss.Color("196")
)

var (
	ActiveBorder = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorAccent)

	InactiveBorder = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorMuted)

	ActiveItem = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorAccent)

	DirItem = lipgloss.NewStyle().
		Foreground(ColorAccent)

	StatusCount = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StatusCountNum = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	StatusKey = lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true)

	StatusLabel = lipgloss.NewStyle().
		Foreground(ColorMuted)

	StatusSep = lipgloss.NewStyle().
		Foreground(ColorMuted)

	StatusMsg = lipgloss.NewStyle().
		Foreground(ColorMatch).
		Bold(true)

	LineNumber = lipgloss.NewStyle().
		Foreground(ColorMuted).
		Width(4).
		Align(lipgloss.Right)

	Comment = lipgloss.NewStyle().
		Foreground(ColorMuted)

	Error = lipgloss.NewStyle().
		Foreground(ColorError)

	Dim = lipgloss.NewStyle().
		Foreground(ColorMuted)

	PanelTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorAccent)

	PickerBorder = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorMuted).
			Padding(0, 1)

	PickerSelected = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorAccent)

	PickerMatch = lipgloss.NewStyle().
		Foreground(ColorMatch)

	PickerFooter = lipgloss.NewStyle().
		Foreground(ColorMuted).
		Italic(true)

	PickerDim = lipgloss.NewStyle().
		Foreground(ColorMuted)

	PickerDivider = lipgloss.NewStyle().
		Foreground(ColorMuted)
)
