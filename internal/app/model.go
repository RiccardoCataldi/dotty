package app

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/riccardo/dotty/internal/fuzzy"
	"github.com/riccardo/dotty/internal/preview"
	"github.com/riccardo/dotty/internal/scan"
	"github.com/riccardo/dotty/internal/tree"
)

type focusPanel int

const (
	focusList focusPanel = iota
	focusPreview
)

type uiMode int

const (
	modeNormal uiMode = iota
	modePicker
)

type tickMsg time.Time

type pickerWarmMsg struct {
	files []fuzzy.Entry
}

type pickerFilterMsg struct {
	gen     int
	query   string
	results []fuzzy.Match
	total   int
}

type Model struct {
	homeDir                 string
	roots                   []*scan.TreeNode
	totalCount              int
	visibleRows             []tree.Row
	expanded                map[string]bool
	cursor                  int
	listOffset              int
	focus                   focusPanel
	mode                    uiMode
	pickerInput             textinput.Model
	pickerAll               []fuzzy.Entry
	pickerWarming           bool
	pickerResults           []fuzzy.Match
	pickerMatchTotal        int
	pickerLastQuery         string
	pickerFilterGen         int
	pickerCursor            int
	pickerOffset            int
	pickerPreview           viewport.Model
	pickerPreviewPath       string
	pickerPreviewTitle      string
	pickerPreviewTotalLines int
	preview                 viewport.Model
	previewTitle            string
	previewTotalLines       int
	width, height           int
	panelHeight             int
	listWidth               int
	previewWidth            int
	listInnerHeight         int
	statusMsg               string
	statusMsgExpiry         time.Time
}

func New(homeDir string, roots []*scan.TreeNode, totalCount int) Model {
	pi := textinput.New()
	pi.Prompt = "> "
	pi.CharLimit = 256
	pi.Width = 60
	ppv := viewport.New(0, 0)
	vp := viewport.New(0, 0)

	m := Model{
		homeDir:       homeDir,
		roots:         roots,
		totalCount:    totalCount,
		expanded:      make(map[string]bool),
		pickerInput:   pi,
		pickerWarming: true,
		pickerPreview: ppv,
		preview:       vp,
		focus:         focusList,
		mode:          modeNormal,
	}
	m.rebuildVisible()
	m.refreshPreview()
	return m
}

func (m *Model) rebuildVisible() {
	m.visibleRows = tree.Visible(m.roots, m.expanded)
	if m.cursor >= len(m.visibleRows) && len(m.visibleRows) > 0 {
		m.cursor = len(m.visibleRows) - 1
	}
	if len(m.visibleRows) == 0 {
		m.cursor = 0
	} else if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *Model) refreshPreview() {
	if len(m.visibleRows) == 0 {
		m.preview.SetContent("")
		m.previewTitle = ""
		return
	}
	if m.cursor >= len(m.visibleRows) {
		m.cursor = len(m.visibleRows) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	entry := m.visibleRows[m.cursor].Node.ToEntry()
	result := preview.Build(entry)
	m.preview.SetContent(result.Content)
	m.previewTitle = result.Title
	m.previewTotalLines = result.TotalLines
	m.preview.GotoTop()
}

func (m *Model) selectedEntry() (scan.Entry, bool) {
	if len(m.visibleRows) == 0 || m.cursor >= len(m.visibleRows) {
		return scan.Entry{}, false
	}
	return m.visibleRows[m.cursor].Node.ToEntry(), true
}

func (m *Model) toggleExpandAtCursor() {
	if m.cursor >= len(m.visibleRows) {
		return
	}
	row := m.visibleRows[m.cursor]
	if !row.Node.IsDir {
		return
	}
	if m.expanded[row.Node.Path] {
		delete(m.expanded, row.Node.Path)
	} else {
		_ = scan.LoadChildren(row.Node)
		m.expanded[row.Node.Path] = true
		m.totalCount = scan.CountNodes(m.roots)
	}
	m.rebuildVisible()
	m.ensureCursorVisible()
}

func (m *Model) collapseAtCursor() {
	if m.cursor >= len(m.visibleRows) {
		return
	}
	row := m.visibleRows[m.cursor]
	if row.Depth > 0 {
		for i := m.cursor - 1; i >= 0; i-- {
			if m.visibleRows[i].Depth < row.Depth {
				m.cursor = i
				delete(m.expanded, m.visibleRows[i].Node.Path)
				m.rebuildVisible()
				m.ensureCursorVisible()
				m.refreshPreview()
				return
			}
		}
	}
	if row.Node.IsDir && m.expanded[row.Node.Path] {
		delete(m.expanded, row.Node.Path)
		m.rebuildVisible()
		m.ensureCursorVisible()
	}
}

func (m *Model) ensureCursorVisible() {
	if m.cursor < m.listOffset {
		m.listOffset = m.cursor
	}
	if m.cursor >= m.listOffset+m.listInnerHeight {
		m.listOffset = m.cursor - m.listInnerHeight + 1
	}
}

func warmPickerCmd(roots []*scan.TreeNode) tea.Cmd {
	return func() tea.Msg {
		return pickerWarmMsg{files: collectAllFiles(roots)}
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) }),
		warmPickerCmd(m.roots),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layoutPanels()
		return m, nil

	case tickMsg:
		if m.statusMsg != "" && time.Now().After(m.statusMsgExpiry) {
			m.statusMsg = ""
		}
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })

	case pickerWarmMsg:
		m.pickerWarming = false
		if m.pickerAll == nil {
			m.pickerAll = msg.files
		}
		if m.mode == modePicker {
			m.pickerFilterGen++
			gen := m.pickerFilterGen
			return m, filterPickerCmd(gen, m.pickerAll, m.pickerInput.Value(), nil, "")
		}
		return m, nil

	case pickerFilterMsg:
		if msg.gen != m.pickerFilterGen {
			return m, nil
		}
		m.applyPickerFilter(msg.query, msg.results, msg.total)
		return m, nil

	case tea.KeyMsg:
		if m.mode == modePicker {
			return m.updatePicker(msg)
		}
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if cmd := m.handleKey(msg); cmd != nil {
			return m, cmd
		}
		return m, nil
	}

	if m.focus == focusPreview && m.mode == modeNormal {
		var cmd tea.Cmd
		m.preview, cmd = m.preview.Update(msg)
		return m, cmd
	}

	return m, nil
}
