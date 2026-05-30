package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/riccardo/dotty/internal/preview"
	"github.com/riccardo/dotty/internal/ui"
)

func (m *Model) layoutPanels() {
	if m.width == 0 || m.height == 0 {
		return
	}
	m.panelHeight = m.height - 1
	m.listWidth = m.width / 3
	m.previewWidth = m.width - m.listWidth
	m.listInnerHeight = max(m.panelHeight-2, 1)
	innerPreviewHeight := max(m.panelHeight-3, 1)
	m.preview.Width = max(m.previewWidth-2, 1)
	m.preview.Height = innerPreviewHeight
	if m.mode == modePicker {
		m.refreshPickerPreview()
	}
}

func limitLines(content string, max int, keepBottom bool) string {
	if max <= 0 {
		return ""
	}
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	if len(lines) <= max {
		return strings.Join(lines, "\n")
	}
	if keepBottom {
		lines = lines[len(lines)-max:]
	} else {
		lines = lines[:max]
	}
	return strings.Join(lines, "\n")
}

func padLines(content string, width, height int) string {
	if width <= 0 || height <= 0 {
		return content
	}
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	for len(lines) < height {
		lines = append(lines, "")
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	for i, line := range lines {
		lines[i] = padLine(line, width)
	}
	return strings.Join(lines, "\n")
}

func padLine(line string, width int) string {
	w := lipgloss.Width(line)
	if w > width {
		return ansi.Truncate(line, width, "")
	}
	if w < width {
		return line + strings.Repeat(" ", width-w)
	}
	return line
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}
	if m.mode == modePicker {
		return padLines(m.renderPickerOverlay(), m.width, m.height)
	}
	panels := lipgloss.JoinHorizontal(lipgloss.Top, m.renderListPanel(), m.renderPreviewPanel())
	status := padLine(m.renderStatusBar(), m.width)
	return padLines(panels+"\n"+status, m.width, m.height)
}

func (m Model) renderListPanel() string {
	border := ui.InactiveBorder
	if m.focus == focusList {
		border = ui.ActiveBorder
	}

	var listLines []string
	end := min(m.listOffset+m.listInnerHeight, len(m.visibleRows))
	for i := m.listOffset; i < end; i++ {
		listLines = append(listLines, preview.RenderListItem(m.visibleRows[i], i == m.cursor))
	}
	for len(listLines) < m.listInnerHeight {
		listLines = append(listLines, "")
	}

	content := strings.Join(listLines, "\n")

	return border.
		Width(max(m.listWidth-2, 1)).
		Height(max(m.panelHeight-2, 1)).
		AlignVertical(lipgloss.Top).
		Render(content)
}

func (m Model) renderPreviewPanel() string {
	border := ui.InactiveBorder
	if m.focus == focusPreview {
		border = ui.ActiveBorder
	}

	title := m.previewTitle
	if m.previewTotalLines > m.preview.Height {
		title = preview.TitleWithPercent(m.previewTitle, m.preview.YOffset, m.preview.Height, m.previewTotalLines)
	}
	titleLine := ui.PanelTitle.Render(" " + title)

	innerWidth := max(m.previewWidth-2, 1)
	titleRendered := padLine(titleLine, innerWidth)
	previewLines := max(m.panelHeight-3, 1)
	content := limitLines(m.preview.View(), previewLines, false)

	panelContent := lipgloss.JoinVertical(lipgloss.Top, titleRendered, content)
	return border.
		Width(max(m.previewWidth-2, 1)).
		Height(max(m.panelHeight-2, 1)).
		AlignVertical(lipgloss.Top).
		Render(panelContent)
}

type statusHintSpec struct {
	key  string
	desc string
}

func formatStatusCount(visible, total int) string {
	return ui.StatusCountNum.Render(fmt.Sprintf("%d", visible)) +
		ui.StatusCount.Render("/") +
		ui.StatusCountNum.Render(fmt.Sprintf("%d", total)) +
		ui.StatusCount.Render(" files")
}

func renderStatusHint(spec statusHintSpec) string {
	return ui.StatusKey.Render(spec.key) + ui.StatusLabel.Render(" "+spec.desc)
}

func renderStatusHints(specs []statusHintSpec) string {
	parts := make([]string, len(specs))
	for i, spec := range specs {
		parts[i] = renderStatusHint(spec)
	}
	return strings.Join(parts, ui.StatusSep.Render(" · "))
}

func (m Model) statusHintSpecs() []statusHintSpec {
	if m.focus == focusPreview {
		return []statusHintSpec{
			{"j/k", "scroll"},
			{"d/u", "page"},
			{"Tab", "list"},
			{"q", "quit"},
		}
	}
	return []statusHintSpec{
		{"j/k", "move"},
		{"Enter", "open"},
		{"/", "find"},
		{"y", "copy path"},
		{"Tab", "preview"},
		{"q", "quit"},
	}
}

func (m Model) renderStatusBar() string {
	left := formatStatusCount(len(m.visibleRows), m.totalCount)

	if m.statusMsg != "" {
		right := ui.StatusMsg.Render(m.statusMsg)
		gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
		if gap < 1 {
			gap = 1
		}
		return left + strings.Repeat(" ", gap) + right
	}

	specs := m.statusHintSpecs()
	right := renderStatusHints(specs)
	for len(specs) > 1 && lipgloss.Width(left)+lipgloss.Width(right) > m.width {
		specs = specs[:len(specs)-1]
		right = renderStatusHints(specs)
	}

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
