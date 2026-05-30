package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
)

type pickerEntry struct {
	RelPath  string
	Path     string
	relLower []rune
}

type pickerMatch struct {
	entry   pickerEntry
	indices []int
}

func newPickerEntry(relPath, path string) pickerEntry {
	return pickerEntry{
		RelPath:  relPath,
		Path:     path,
		relLower: []rune(strings.ToLower(relPath)),
	}
}

func collectAllFiles(roots []*TreeNode) []pickerEntry {
	var files []pickerEntry
	var walk func(node *TreeNode)
	walk = func(node *TreeNode) {
		if !node.IsDir {
			files = append(files, newPickerEntry(node.RelPath, node.Path))
			return
		}
		_ = loadChildren(node)
		for _, child := range node.Children {
			walk(child)
		}
	}
	for _, root := range roots {
		walk(root)
	}
	return files
}

func (m *model) openPicker() {
	if m.pickerAll == nil {
		m.pickerAll = collectAllFiles(m.roots)
	}
	m.pickerInput.SetValue("")
	m.pickerInput.CursorEnd()
	m.pickerInput.Focus()
	m.pickerCursor = 0
	m.pickerOffset = 0
	m.pickerLastQuery = ""
	m.filterPicker()
	m.mode = modePicker
}

func (m *model) closePicker() {
	m.mode = modeNormal
	m.pickerPreviewPath = ""
	m.pickerInput.Blur()
}

func (m *model) filterPicker() {
	query := m.pickerInput.Value()
	results, total := filterAndRankPickerFiles(m.pickerAll, query, m.pickerResults, m.pickerLastQuery)
	m.applyPickerFilter(query, results, total)
}

func (m *model) applyPickerFilter(query string, results []pickerMatch, total int) {
	if query != m.pickerLastQuery {
		m.pickerCursor = 0
		m.pickerOffset = 0
	}
	m.pickerLastQuery = query
	m.pickerResults = results
	m.pickerMatchTotal = total
	if m.pickerCursor >= len(m.pickerResults) {
		if len(m.pickerResults) > 0 {
			m.pickerCursor = len(m.pickerResults) - 1
		} else {
			m.pickerCursor = 0
		}
	}
	m.ensurePickerCursorVisible()
	m.refreshPickerPreview()
}

func filterPickerCmd(gen int, all []pickerEntry, query string, prev []pickerMatch, prevQuery string) tea.Cmd {
	return func() tea.Msg {
		results, total := filterAndRankPickerFiles(all, query, prev, prevQuery)
		return pickerFilterMsg{gen: gen, query: query, results: results, total: total}
	}
}

func (m *model) refreshPickerPreview() {
	_, _, _, previewColW, listHeight := m.pickerLayout()
	bodyH := max(listHeight-1, 1)
	m.pickerPreview.Width = max(previewColW, 1)
	m.pickerPreview.Height = bodyH
	m.pickerPreview.Style = lipgloss.NewStyle().Width(previewColW).Height(bodyH)

	if len(m.pickerResults) == 0 || m.pickerCursor >= len(m.pickerResults) {
		m.pickerPreviewPath = ""
		m.pickerPreview.SetContent(stylePickerDim.Render("No file selected"))
		m.pickerPreviewTitle = ""
		m.pickerPreviewTotalLines = 0
		return
	}

	e := m.pickerResults[m.pickerCursor].entry
	if e.Path == m.pickerPreviewPath {
		return
	}
	m.pickerPreviewPath = e.Path

	name := filepath.Base(strings.TrimSuffix(e.RelPath, "/"))
	entry := Entry{Name: name, Path: e.Path, IsDir: false}
	result := buildPreview(entry)
	m.pickerPreview.SetContent(clampPreviewToWidth(result.Content, previewColW))
	m.pickerPreviewTitle = result.Title
	m.pickerPreviewTotalLines = result.TotalLines
	m.pickerPreview.GotoTop()
}

func clampPreviewToWidth(content string, width int) string {
	if width < 4 || content == "" {
		return content
	}
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if runewidth.StringWidth(ansi.Strip(line)) > width {
			lines[i] = ansi.Truncate(line, width, "…")
		}
	}
	return strings.Join(lines, "\n")
}

func (m *model) pickerLayout() (pickerWidth, innerW, listColW, previewColW, listHeight int) {
	pickerWidth = max(min(m.width*9/10, m.width-4), 60)
	innerW = pickerWidth - 4
	listColW = innerW * 11 / 20
	if listColW < 20 {
		listColW = innerW / 2
	}
	previewColW = innerW - listColW - 1
	if previewColW < 16 {
		previewColW = 16
		listColW = innerW - previewColW - 1
	}

	listHeight = max(min(m.height*2/3, m.height-8), 10)
	if listHeight > m.height-6 {
		listHeight = max(m.height-6, 6)
	}
	return pickerWidth, innerW, listColW, previewColW, listHeight
}

func (m *model) pickerResultsHeight() int {
	_, _, _, _, h := m.pickerLayout()
	return h
}

func (m *model) ensurePickerCursorVisible() {
	h := m.pickerResultsHeight()
	if m.pickerCursor < m.pickerOffset {
		m.pickerOffset = m.pickerCursor
	}
	if m.pickerCursor >= m.pickerOffset+h {
		m.pickerOffset = m.pickerCursor - h + 1
	}
}

func (m *model) confirmPicker() {
	if len(m.pickerResults) == 0 || m.pickerCursor >= len(m.pickerResults) {
		m.closePicker()
		return
	}
	selected := m.pickerResults[m.pickerCursor].entry
	m.closePicker()
	m.revealRelPath(selected.RelPath)
	m.refreshPreview()
}

func (m *model) revealRelPath(target string) {
	for _, root := range m.roots {
		if revealFrom(m, root, target) {
			break
		}
	}
	m.totalCount = countTreeNodes(m.roots)
	m.invalidatePickerCache()
	m.rebuildVisible()
	if idx := findRowByRelPath(m.visibleRows, target); idx >= 0 {
		m.cursor = idx
		m.ensureCursorVisible()
	}
}

func revealFrom(m *model, node *TreeNode, target string) bool {
	if node.RelPath == target {
		return true
	}
	if !node.IsDir {
		return false
	}
	if !strings.HasPrefix(target, node.RelPath) {
		return false
	}
	_ = loadChildren(node)
	m.expanded[node.Path] = true
	for _, child := range node.Children {
		if child.RelPath == target {
			return true
		}
		if child.IsDir && strings.HasPrefix(target, child.RelPath) {
			if revealFrom(m, child, target) {
				return true
			}
		}
	}
	return false
}

func (m *model) invalidatePickerCache() {
	m.pickerAll = nil
}

func (m model) updatePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "esc" || msg.String() == "q":
		m.closePicker()
		return m, nil
	case key.Matches(msg, keys.Down):
		if m.pickerCursor < len(m.pickerResults)-1 {
			m.pickerCursor++
			m.ensurePickerCursorVisible()
			m.refreshPickerPreview()
		}
		return m, nil
	case key.Matches(msg, keys.Up):
		if m.pickerCursor > 0 {
			m.pickerCursor--
			m.ensurePickerCursorVisible()
			m.refreshPickerPreview()
		}
		return m, nil
	case key.Matches(msg, keys.ScrollDown):
		m.pickerPreview.LineDown(1)
		return m, nil
	case key.Matches(msg, keys.ScrollUp):
		m.pickerPreview.LineUp(1)
		return m, nil
	case msg.String() == "enter":
		m.confirmPicker()
		return m, nil
	}

	var cmd tea.Cmd
	oldQuery := m.pickerInput.Value()
	m.pickerInput, cmd = m.pickerInput.Update(msg)
	newQuery := m.pickerInput.Value()
	if newQuery != oldQuery {
		if newQuery != m.pickerLastQuery {
			m.pickerCursor = 0
			m.pickerOffset = 0
		}
		m.pickerFilterGen++
		gen := m.pickerFilterGen
		prev := m.pickerResults
		prevQ := m.pickerLastQuery
		return m, tea.Batch(cmd, filterPickerCmd(gen, m.pickerAll, newQuery, prev, prevQ))
	}
	return m, cmd
}

func renderPickerItem(match pickerMatch, selected bool, listColW int) string {
	entry := match.entry
	prefix := "  "
	if selected {
		prefix = "> "
	}
	// account for prefix width
	displayW := max(listColW-runewidth.StringWidth(prefix), 8)
	display := pickerDisplayPath(entry.RelPath, displayW)
	indices := mapIndicesToDisplay(entry.RelPath, display, match.indices)
	text := highlightPickerMatches(display, indices)
	line := prefix + text
	if selected {
		return padLine(stylePickerSelected.Render(line), listColW)
	}
	return padLine(line, listColW)
}

func highlightPickerMatches(text string, indices []int) string {
	if len(indices) == 0 {
		return text
	}
	matchSet := make(map[int]bool, len(indices))
	for _, i := range indices {
		matchSet[i] = true
	}
	var b strings.Builder
	for i := 0; i < len(text); {
		if matchSet[i] {
			r, size := utf8.DecodeRuneInString(text[i:])
			b.WriteString(stylePickerMatch.Render(string(r)))
			i += size
		} else {
			r, size := utf8.DecodeRuneInString(text[i:])
			b.WriteString(string(r))
			i += size
		}
	}
	return b.String()
}

func (m model) renderPickerOverlay() string {
	pickerWidth, innerW, listColW, previewColW, listHeight := m.pickerLayout()

	m.pickerInput.Width = innerW
	promptLine := padLine(m.pickerInput.View(), innerW)
	sep := padLine(strings.Repeat("─", max(innerW, 1)), innerW)

	query := m.pickerInput.Value()
	var lines []string
	end := min(m.pickerOffset+listHeight, len(m.pickerResults))
	for i := m.pickerOffset; i < end; i++ {
		lines = append(lines, renderPickerItem(
			m.pickerResults[i],
			i == m.pickerCursor,
			listColW,
		))
	}
	for len(lines) < listHeight {
		lines = append(lines, "")
	}

	listBlock := lipgloss.NewStyle().Width(listColW).Height(listHeight).Render(strings.Join(lines, "\n"))

	previewTitle := m.pickerPreviewTitle
	if m.pickerPreviewTotalLines > m.pickerPreview.Height {
		previewTitle = previewTitleWithPercent(m.pickerPreviewTitle, m.pickerPreview.YOffset, m.pickerPreview.Height, m.pickerPreviewTotalLines)
	}
	titleLine := padLine(stylePanelTitle.Render(" "+previewTitle), previewColW)
	previewBody := m.pickerPreview.View()
	if previewBody == "" {
		previewBody = stylePickerDim.Render("No file selected")
	}
	previewCol := lipgloss.NewStyle().Width(previewColW).Height(listHeight).Render(
		lipgloss.JoinVertical(lipgloss.Top, titleLine, previewBody),
	)

	divider := lipgloss.NewStyle().Width(1).Height(listHeight).Render(
		stylePickerDivider.Render("│"),
	)
	split := lipgloss.JoinHorizontal(lipgloss.Top, listBlock, divider, previewCol)

	matchTotal := m.pickerMatchTotal
	if query == "" {
		matchTotal = len(m.pickerAll)
	}
	footerText := stylePickerFooter.Render(
		formatPickerFooter(m.pickerCursor, len(m.pickerResults), matchTotal, len(m.pickerAll)),
	)

	body := lipgloss.JoinVertical(lipgloss.Left, promptLine, sep, split, padLine(footerText, innerW))
	modal := stylePickerBorder.Width(pickerWidth).Render(body)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}

func formatPickerFooter(cursor, shown, matchTotal, cacheTotal int) string {
	if shown == 0 {
		return "  no matches"
	}
	idx := cursor + 1
	if matchTotal == shown && matchTotal == cacheTotal {
		return fmt.Sprintf("  %d/%d", idx, shown)
	}
	if matchTotal == shown {
		return fmt.Sprintf("  %d/%d  (%d total)", idx, shown, cacheTotal)
	}
	return fmt.Sprintf("  %d/%d  (%d matches, %d total)", idx, shown, matchTotal, cacheTotal)
}
